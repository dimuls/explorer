package ent

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
)

type Peer struct {
	ID  int64  `json:"id" db:"id" goqu:"skipinsert"`
	URL string `json:"url" db:"url"`
}

type Channel struct {
	ID string `json:"id" db:"id"`
}

type PeerChannel struct {
	PeerID    int64  `json:"peer_id" db:"peer_id"`
	ChannelID string `json:"channel_id" db:"channel_id"`
}

type CommonConfig common.Config

func (cc CommonConfig) Value() (driver.Value, error) {
	ccJSON, err := json.Marshal(&cc)
	if err != nil {
		return nil, err
	}
	return driver.Value(ccJSON), nil
}

func (cc *CommonConfig) Scan(src interface{}) error {
	ccJSON, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unable to cast src to byte array")
	}
	err := json.Unmarshal(ccJSON, cc)
	if err != nil {
		return fmt.Errorf("unmarshal common config JSON: %w", err)
	}
	return nil
}

type ChannelConfig struct {
	ID        int64         `json:"id" db:"id" goqu:"skipinsert"`
	ChannelID string        `json:"channel_id" db:"channel_id"`
	Raw       []byte        `json:"raw" db:"raw"`
	Parsed    *CommonConfig `json:"parsed" db:"parsed"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}

type Chaincode struct {
	ID      int64  `json:"id" db:"id" goqu:"skipinsert"`
	Name    string `json:"name" db:"name"`
	Version string `json:"version" db:"version"`
}

type ChannelChaincode struct {
	ChannelID   string `json:"channel_id" db:"channel_id"`
	ChaincodeID int64  `json:"chaincode_id" db:"chaincode_id"`
}

type Block struct {
	ID        int64  `json:"id" db:"id" goqu:"skipinsert"`
	ChannelID string `json:"channel_id" db:"channel_id"`
	Number    int64  `json:"number" db:"number"`
}

type Transaction struct {
	ID        string    `json:"id" db:"id"`
	BlockID   int64     `json:"block_id" db:"block_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type State struct {
	Key           string `json:"key" db:"key"`
	TransactionID string `json:"transaction_id" db:"transaction_id"`
	RawValue      []byte `json:"raw_value" db:"raw_value"`
	Value         []byte `json:"value" db:"value"`
}

type OldState struct {
	ID            int64  `json:"id" db:"id" goqu:"skipinsert"`
	TransactionID string `json:"transaction_id" db:"transaction_id"`
	Key           string `json:"key" db:"key"`
	RawValue      []byte `json:"raw_value" db:"raw_value"`
	Value         []byte `json:"value" db:"value"`
}
