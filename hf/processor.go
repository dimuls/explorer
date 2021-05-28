package hf

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/sirupsen/logrus"

	"explorer/ent"
	"explorer/pg"
)

type ProcessorConfig struct {
	ChannelID        string   `yaml:"channel_id"`
	Chaincodes       []string `yaml:"chaincodes"`
	Organization     string   `yaml:"organization"`
	User             string   `yaml:"user"`
	FabricConfigFile string   `yaml:"fabric_config_file"`
}

type Processor struct {
	channelID string
	storage   *pg.Storage
	log       *logrus.Entry
	wg        sync.WaitGroup
	close     chan struct{}
}

type chaincodeEventSource struct {
	reg fab.Registration
	es  <-chan *fab.CCEvent
}

func init() {
	spew.Config.DisableMethods = true
}

func NewProcessor(c ProcessorConfig, s *pg.Storage) (p *Processor, err error) {

	log := logrus.WithFields(logrus.Fields{
		"subsystem":  "processor",
		"channel_id": c.ChannelID,
	})

	lastBlockID, err := s.LastBlockID(context.TODO())
	if err != nil {
		return nil, fmt.Errorf(
			"get last block ID from storage: %w", err)
	}

	log.WithField("last_block_id", lastBlockID).Debug("got last block ID")

	fsdk, err := fabsdk.New(config.FromFile(c.FabricConfigFile))
	if err != nil {
		return nil, fmt.Errorf("create fabric SDK: %w", err)
	}

	chCtx := fsdk.ChannelContext(c.ChannelID,
		fabsdk.WithUser(c.User))

	evClient, err := event.New(chCtx, event.WithBlockEvents(), event.WithSeekType(seek.FromBlock),
		event.WithBlockNum(uint64(lastBlockID)))
	if err != nil {
		return nil, fmt.Errorf("create event client: %w", err)
	}

	var (
		bsEsReg fab.Registration
		bsEs    <-chan *fab.BlockEvent
	)

	bsEsReg, bsEs, err = evClient.RegisterBlockEvent()
	if err != nil {
		return nil, fmt.Errorf("register block event: %w", err)
	}
	defer func() {
		if err != nil {
			evClient.Unregister(bsEsReg)
		}
	}()

	//var ccEvSrcs []chaincodeEventSource
	//defer func() {
	//	if err != nil {
	//		for _, cc := range ccEvSrcs {
	//			evClient.Unregister(cc.reg)
	//		}
	//	}
	//}()

	//for _, cc := range c.Chaincodes {
	//	reg, es, err := evClient.RegisterChaincodeEvent(cc, "")
	//	if err != nil {
	//		return nil, err
	//	}
	//	ccEvSrcs = append(ccEvSrcs, chaincodeEventSource{
	//		reg: reg,
	//		es:  es,
	//	})
	//}

	p = &Processor{
		channelID: c.ChannelID,
		storage:   s,
		log:       log,
		close:     make(chan struct{}),
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer evClient.Unregister(bsEsReg)

		for {
			select {
			case <-p.close:
				return
			case be := <-bsEs:
				for {
					log := p.log.WithFields(logrus.Fields{
						"block_number": be.Block.Header.Number,
						"peer_url":     be.SourceURL,
					})

					log.Debug("block received")

					err := p.processBlockEvent(log, be)
					if err == nil {
						log.Info("block processed")
						break
					}
					p.log.WithError(err).Error("failed to process block")

					t := time.NewTimer(10 * time.Second)

					select {
					case <-p.close:
						t.Stop()
						return
					case <-t.C:
					}
				}
			}
		}
	}()

	//for _, ccEvSrc := range ccEvSrcs {
	//	p.wg.Add(1)
	//	go func(ccEvSrc chaincodeEventSource) {
	//		defer p.wg.Done()
	//		defer client.Unregister(ccEvSrc.reg)
	//
	//		for {
	//			select {
	//			case <-p.close:
	//				return
	//			case ce := <-ccEvSrc.es:
	//				p.processChaincodeEvent(ce)
	//			}
	//		}
	//	}(ccEvSrc)
	//}

	return p, nil
}

func (p *Processor) Close() {
	close(p.close)
	p.wg.Wait()
}

const lifecycle = "_lifecycle"

func (p *Processor) processBlockEvent(log *logrus.Entry, be *fab.BlockEvent) error {

	var (
		peer           ent.Peer
		channel        ent.Channel
		channelConfigs []ent.ChannelConfig
		chaincodes     []ent.Chaincode
		block          ent.Block
		transactions   []ent.Transaction
		states         []ent.State
	)

	peer.URL = be.SourceURL
	channel.ID = p.channelID
	block.ChannelID = p.channelID

	if be.Block.Header.Number > uint64(math.MaxInt64) {
		return fmt.Errorf("block number greater than max int64")
	}

	block.Number = int64(be.Block.Header.Number)

	for _, d := range be.Block.Data.Data {

		envelope := &common.Envelope{}
		err := proto.Unmarshal(d, envelope)
		if err != nil {
			return fmt.Errorf("unmarshal envelope: %w", err)
		}

		payload := &common.Payload{}
		err = proto.Unmarshal(envelope.Payload, payload)
		if err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		channelHeader := &common.ChannelHeader{}
		err = proto.Unmarshal(payload.Header.ChannelHeader, channelHeader)
		if err != nil {
			return fmt.Errorf("unmarshal channel header: %w", err)
		}

		var transaction ent.Transaction

		transaction.ID = channelHeader.TxId
		transaction.BlockID = block.ID

		transactions = append(transactions, transaction)

		log.WithField("transaction_type",
			common.HeaderType(channelHeader.Type)).
			Debug("transaction found")

		headerType := common.HeaderType(channelHeader.Type)

		switch headerType {

		case common.HeaderType_MESSAGE,
			common.HeaderType_CONFIG_UPDATE,
			common.HeaderType_ORDERER_TRANSACTION,
			common.HeaderType_DELIVER_SEEK_INFO,
			common.HeaderType_CHAINCODE_PACKAGE:

			return fmt.Errorf(
				"transactions of type `%s` not implemented", headerType)

		case common.HeaderType_CONFIG:

			var channelConfig ent.ChannelConfig

			channelConfig.ChannelID = p.channelID
			channelConfig.CreatedAt = channelHeader.Timestamp.AsTime()
			channelConfig.Raw = payload.Data

			configEnvelope := &common.ConfigEnvelope{}

			err := proto.Unmarshal(payload.Data, configEnvelope)
			if err != nil {
				return fmt.Errorf("parse channel config: %w", err)
			}

			commonConfig := ent.CommonConfig(*configEnvelope.Config)
			channelConfig.Parsed = &commonConfig

			channelConfigs = append(channelConfigs, channelConfig)

		case common.HeaderType_ENDORSER_TRANSACTION:

			channelHeaderExtension := &fabricPeer.ChaincodeHeaderExtension{}

			err = proto.Unmarshal(channelHeader.Extension, channelHeaderExtension)
			if err != nil {
				return fmt.Errorf(
					"unmarshal channel header extension: %w", err)
			}

			var chaincode ent.Chaincode

			chaincode.Name = channelHeaderExtension.ChaincodeId.Name
			chaincode.Version = channelHeaderExtension.ChaincodeId.Version

			chaincodes = append(chaincodes, chaincode)

			fabricTransaction := &fabricPeer.Transaction{}

			err = proto.Unmarshal(payload.Data, fabricTransaction)
			if err != nil {
				return fmt.Errorf("unmarshal transaction: %w", err)
			}

			for _, a := range fabricTransaction.Actions {
				chaincodeActionPayload := &fabricPeer.ChaincodeActionPayload{}

				err = proto.Unmarshal(a.Payload, chaincodeActionPayload)
				if err != nil {
					return fmt.Errorf(
						"unmarshal chaincode action payload: %w", err)
				}

				//chaincodeProposalPayload := &fabricPeer.ChaincodeProposalPayload{}
				//
				//err = proto.Unmarshal(
				//	chaincodeActionPayload.ChaincodeProposalPayload,
				//	chaincodeProposalPayload)
				//if err != nil {
				//	return fmt.Errorf(
				//		"unmarshal chaincode proposal payload: %w", err)
				//}
				//
				//chaincodeInvocationSpec := &fabricPeer.ChaincodeInvocationSpec{}
				//
				//err = proto.Unmarshal(
				//	chaincodeProposalPayload.Input,
				//	chaincodeInvocationSpec)
				//if err != nil {
				//	return fmt.Errorf(
				//		"unmarshal chaincode invocation spec: %w", err)
				//}
				//
				//if chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId.Name == lifecycle {
				//	p.log.WithField("chaincode_id", lifecycle).
				//		Warning("transaction of unimplemented chaincode found")
				//	continue
				//}

				proposalResponsePayload := &fabricPeer.ProposalResponsePayload{}

				err = proto.Unmarshal(
					chaincodeActionPayload.Action.ProposalResponsePayload,
					proposalResponsePayload)
				if err != nil {
					return fmt.Errorf(
						"unmarshal proposal response payload: %w", err)
				}

				chaincodeAction := &fabricPeer.ChaincodeAction{}
				err = proto.Unmarshal(
					proposalResponsePayload.Extension,
					chaincodeAction)
				if err != nil {
					return fmt.Errorf(
						"unmarshal chaincode action: %w", err)
				}

				txReadWriteSet := &rwset.TxReadWriteSet{}
				err = proto.Unmarshal(
					chaincodeAction.Results,
					txReadWriteSet)
				if err != nil {
					return fmt.Errorf(
						"unmarshal transaction read write set: %w", err)
				}

				for _, rw := range txReadWriteSet.NsRwset {
					kvRWSet := &kvrwset.KVRWSet{}
					err = proto.Unmarshal(rw.Rwset, kvRWSet)
					if err != nil {
						return fmt.Errorf(
							"unmarshal kv rw set: %w", err)
					}

					for _, w := range kvRWSet.Writes {
						parsedValue, err := parseValue(chaincode.Name, w.Key, w.Value)
						if err != nil {
							fmt.Errorf("parse value: %w", err)
						}

						states = append(states, ent.State{
							Key:           w.Key,
							TransactionID: transaction.ID,
							RawValue:      w.Value,
							Value:         parsedValue,
						})
					}

				}
			}
		}
	}

	ctx := context.TODO()

	tx, err := p.storage.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction in storage: %w", err)
	}

	peer.ID, err = p.storage.AddPeerTx(ctx, tx, peer)
	if err != nil {
		return fmt.Errorf("add peer to storage: %w", err)
	}

	err = p.storage.AddChannelTx(ctx, tx, channel)
	if err != nil {
		return fmt.Errorf("add channel to storage: %w", err)
	}

	err = p.storage.AddPeerChannelTx(ctx, tx, ent.PeerChannel{
		PeerID:    peer.ID,
		ChannelID: channel.ID,
	})
	if err != nil {
		return fmt.Errorf("add peer_channel to storage: %w", err)
	}

	for _, cc := range channelConfigs {
		err = p.storage.AddChannelConfigTx(ctx, tx, cc)
		if err != nil {
			return fmt.Errorf("add channel config to storage: %w", err)
		}
	}

	for _, c := range chaincodes {
		c.ID, err = p.storage.AddChaincodeTx(ctx, tx, c)
		if err != nil {
			return fmt.Errorf("add chaincode to storage: %w", err)
		}

		err = p.storage.AddChannelChaincodeTx(ctx, tx, ent.ChannelChaincode{
			ChannelID:   channel.ID,
			ChaincodeID: c.ID,
		})
		if err != nil {
			return fmt.Errorf("add channel_chaincode to storage: %w", err)
		}
	}

	block.ID, err = p.storage.AddBlockTx(ctx, tx, block)
	if err != nil {
		return fmt.Errorf("add block to storage: %w", err)
	}

	for _, t := range transactions {
		t.BlockID = block.ID
		err = p.storage.AddTransactionTx(ctx, tx, t)
		if err != nil {
			return fmt.Errorf("add transaction to storage: %w", err)
		}
	}

	for _, s := range states {
		err = p.storage.AddStateTx(ctx, tx, s)
		if err != nil {
			return fmt.Errorf("add state to storage: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

//func (p *Processor) processChaincodeEvent(ce *fab.CCEvent) error {
//	// TODO
//	return nil
//}
