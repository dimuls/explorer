package hf

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dimuls/fabric-sdk-go/pkg/client/event"
	"github.com/dimuls/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/dimuls/fabric-sdk-go/pkg/core/config"
	"github.com/dimuls/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/dimuls/fabric-sdk-go/pkg/fabsdk"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	fabricPeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/sirupsen/logrus"

	"explorer"
)

type ProcessorConfig struct {
	ChannelName      string   `yaml:"channel_name"`
	Chaincodes       []string `yaml:"chaincodes"`
	Organization     string   `yaml:"organization"`
	User             string   `yaml:"user"`
	FabricConfigFile string   `yaml:"fabric_config_file"`
}

type Processor struct {
	channelName string
	storage     Storage
	log         *logrus.Entry
	wg          sync.WaitGroup
	close       chan struct{}
}

type chaincodeEventSource struct {
	reg fab.Registration
	es  <-chan *fab.CCEvent
}

func init() {
	spew.Config.DisableMethods = true
}

func NewProcessor(c ProcessorConfig, s Storage) (p *Processor, err error) {

	log := logrus.WithFields(logrus.Fields{
		"subsystem":  "processor",
		"channel_id": c.ChannelName,
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

	chCtx := fsdk.ChannelContext(c.ChannelName,
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
		channelName: c.ChannelName,
		storage:     s,
		log:         log,
		close:       make(chan struct{}),
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
		peer           = &explorer.Peer{}
		channel        = &explorer.Channel{}
		channelConfigs []*explorer.ChannelConfig
		chaincodes     []*explorer.Chaincode
		block          = &explorer.Block{}
		transactions   []*explorer.Transaction
		states         []*explorer.State
	)

	if be.Block.Header.Number > uint64(math.MaxInt64) {
		return fmt.Errorf("block number greater than max int64")
	}

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

		transaction := &explorer.Transaction{}

		transaction.Id = channelHeader.TxId
		transaction.BlockId = block.Id
		transaction.CreatedAt = channelHeader.Timestamp

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

			channelConfig := &explorer.ChannelConfig{}

			channelConfig.CreatedAt = channelHeader.Timestamp
			channelConfig.Raw = payload.Data

			configEnvelope := &common.ConfigEnvelope{}

			err := proto.Unmarshal(payload.Data, configEnvelope)
			if err != nil {
				return fmt.Errorf("parse channel config: %w", err)
			}

			channelConfig.Parsed, err = json.Marshal(configEnvelope.Config)
			if err != nil {
				return fmt.Errorf("JSON marshal config: %w", err)
			}

			channelConfigs = append(channelConfigs, channelConfig)

		case common.HeaderType_ENDORSER_TRANSACTION:

			channelHeaderExtension := &fabricPeer.ChaincodeHeaderExtension{}

			err = proto.Unmarshal(channelHeader.Extension, channelHeaderExtension)
			if err != nil {
				return fmt.Errorf(
					"unmarshal channel header extension: %w", err)
			}

			chaincode := &explorer.Chaincode{}

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
						stateType, parsedValue, err := parseValue(chaincode.Name, w.Key, w.Value)
						if err != nil {
							fmt.Errorf("parse value: %w", err)
						}

						states = append(states, &explorer.State{
							Key:           w.Key,
							CreatedAt:     transaction.CreatedAt,
							Type:          stateType,
							TransactionId: transaction.Id,
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

	peer.Url = be.SourceURL
	peer.Id, err = p.storage.AddPeerTx(ctx, tx, peer)
	if err != nil {
		return fmt.Errorf("add peer to storage: %w", err)
	}

	channel.Name = p.channelName
	channel.Id, err = p.storage.AddChannelTx(ctx, tx, channel)
	if err != nil {
		return fmt.Errorf("add channel to storage: %w", err)
	}

	err = p.storage.AddPeerChannelTx(ctx, tx, &explorer.PeerChannel{
		PeerId:    peer.Id,
		ChannelId: channel.Id,
	})
	if err != nil {
		return fmt.Errorf("add peer_channel to storage: %w", err)
	}

	for _, cc := range channelConfigs {
		cc.ChannelId = channel.Id
		err = p.storage.AddChannelConfigTx(ctx, tx, cc)
		if err != nil {
			return fmt.Errorf("add channel config to storage: %w", err)
		}
	}

	for _, c := range chaincodes {
		c.Id, err = p.storage.AddChaincodeTx(ctx, tx, c)
		if err != nil {
			return fmt.Errorf("add chaincode to storage: %w", err)
		}

		err = p.storage.AddChannelChaincodeTx(ctx, tx, &explorer.ChannelChaincode{
			ChannelId:   channel.Id,
			ChaincodeId: c.Id,
		})
		if err != nil {
			return fmt.Errorf("add channel_chaincode to storage: %w", err)
		}
	}

	block.Number = int64(be.Block.Header.Number)
	block.ChannelId = channel.Id
	block.Id, err = p.storage.AddBlockTx(ctx, tx, block)
	if err != nil {
		return fmt.Errorf("add block to storage: %w", err)
	}

	for _, t := range transactions {
		t.ChannelId = channel.Id
		t.BlockId = block.Id
		err = p.storage.AddTransactionTx(ctx, tx, t)
		if err != nil {
			return fmt.Errorf("add transaction to storage: %w", err)
		}
	}

	for _, s := range states {
		s.ChannelId = channel.Id
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
