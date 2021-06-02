package hf

import (
	"context"
	"database/sql"
	"explorer"
)

type Storage interface {
	LastBlockID(ctx context.Context) (lastID int64, err error)
	BeginTx(ctx context.Context) (*sql.Tx, error)
	AddPeerTx(ctx context.Context, tx *sql.Tx, p *explorer.Peer) (id int64, err error)
	AddChannelTx(ctx context.Context, tx *sql.Tx, c *explorer.Channel) (int64, error)
	AddChannelConfigTx(ctx context.Context, tx *sql.Tx, cc *explorer.ChannelConfig) error
	AddPeerChannelTx(ctx context.Context, tx *sql.Tx, c *explorer.PeerChannel) error
	AddChaincodeTx(ctx context.Context, tx *sql.Tx, c *explorer.Chaincode) (id int64, err error)
	AddChannelChaincodeTx(ctx context.Context, tx *sql.Tx, cc *explorer.ChannelChaincode) error
	AddBlockTx(ctx context.Context, tx *sql.Tx, b *explorer.Block) (id int64, err error)
	AddTransactionTx(ctx context.Context, tx *sql.Tx, t *explorer.Transaction) error
	AddStateTx(ctx context.Context, tx *sql.Tx, as *explorer.State) (err error)
}
