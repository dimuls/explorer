package pg

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/doug-martin/goqu/v9"

	"explorer"
)

func (e *Explorer) LastBlockID(ctx context.Context) (lastID int64, err error) {
	found, err := e.db.
		From("block").
		Select(goqu.COALESCE(goqu.MAX("id"), 0)).
		ScanValContext(ctx, &lastID)
	if err != nil {
		return
	}
	if !found {
		lastID = 0
	}
	return
}

const (
	postgresDialect = "postgres"

	peer             = "peer"
	channel          = "channel"
	peerChannel      = "peer_channel"
	channelConfig    = "channel_config"
	chaincode        = "chaincode"
	channelChaincode = "channel_chaincode"
	block            = "block"
	transaction      = "transaction"
	state            = "state"
	oldState         = "old_state"
)

func (e *Explorer) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return e.db.Db.BeginTx(ctx, nil)
}

func (e *Explorer) AddPeerTx(ctx context.Context, tx *sql.Tx,
	p *explorer.Peer) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				e.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.Select("id").
		From(peer).
		Where(goqu.Ex{"url": p.Url}).
		ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return id, nil
	}

	_, err = txx.
		Insert(peer).
		Rows(p).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add peer to DB: %w", err)
	}

	return id, nil
}

func (e *Explorer) AddChannelTx(ctx context.Context, tx *sql.Tx,
	c *explorer.Channel) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				e.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.Select("id").
		From(channel).
		Where(goqu.Ex{"name": c.Name}).
		ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("get channel from DB: %w", err)
	}
	if exists {
		return id, nil
	}

	_, err = txx.
		Insert(channel).
		Rows(c).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add channel to DB: %w", err)
	}

	return id, nil
}

func (e *Explorer) AddChannelConfigTx(ctx context.Context, tx *sql.Tx,
	cc *explorer.ChannelConfig) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(channelConfig).
		Rows(goqu.Record{
			"channel_id": cc.ChannelId,
			"raw":        hex.EncodeToString(cc.Raw),
			"parsed":     cc.Parsed,
			"created_at": cc.CreatedAt.AsTime()}).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			e.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (e *Explorer) AddPeerChannelTx(ctx context.Context, tx *sql.Tx,
	c *explorer.PeerChannel) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(peerChannel).
		Rows(c).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			e.log.WithError(err2).Error("failed to rollback transaction")
		}
	}

	return err
}

func (e *Explorer) AddChaincodeTx(ctx context.Context, tx *sql.Tx,
	c *explorer.Chaincode) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				e.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.
		Select("id").
		From(chaincode).
		Where(goqu.Ex{"name": c.Name, "version": c.Version}).
		ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return id, nil
	}

	_, err = txx.
		Insert(chaincode).
		Rows(c).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add chaincode to DB: %w", err)
	}

	return id, nil
}

func (e *Explorer) AddChannelChaincodeTx(ctx context.Context, tx *sql.Tx,
	cc *explorer.ChannelChaincode) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(channelChaincode).
		Rows(cc).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			e.log.WithError(err2).Error("failed to rollback transaction")
		}
	}

	return err
}

func (e *Explorer) AddBlockTx(ctx context.Context, tx *sql.Tx,
	b *explorer.Block) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				e.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.
		Select("id").
		From(block).
		Where(goqu.Ex{"channel_id": b.ChannelId, "number": b.Number}).
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return id, nil
	}

	fmt.Println(b)

	_, err = txx.
		Insert(block).
		Rows(b).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add peer to DB: %w", err)
	}

	return id, nil
}

func (e *Explorer) AddTransactionTx(ctx context.Context, tx *sql.Tx,
	t *explorer.Transaction) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(transaction).
		Rows(goqu.Record{
			"id":         t.Id,
			"channel_id": t.ChannelId,
			"block_id":   t.BlockId,
			"created_at": t.CreatedAt.AsTime(),
		}).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			e.log.WithError(err2).Error("failed to rollback transaction")
		}
	}

	return err
}

func (e *Explorer) AddStateTx(ctx context.Context, tx *sql.Tx,
	as *explorer.State) (err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				e.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	var os explorer.State

	rows, err := txx.Select("key", "channel_id", "transaction_id",
		"type", "raw_value", "value", "created_at").
		From(state).
		Where(goqu.Ex{"key": as.Key}).
		Executor().Query()
	if err != nil {
		return fmt.Errorf("get actual state: %w", err)
	}

	defer rows.Close()

	var exists bool
	if rows.Next() {
		exists = true
		var createdAt time.Time
		err = rows.Scan(&os.Key, &os.ChannelId, &os.TransactionId, &os.Type,
			&os.RawValue, &os.Value, &createdAt)
		if err != nil {
			return err
		}
		os.CreatedAt = timestamppb.New(createdAt)
	}

	ar := goqu.Record{
		"key":            as.Key,
		"channel_id":     as.ChannelId,
		"transaction_id": as.TransactionId,
		"type":           as.Type,
		"raw_value":      hex.EncodeToString(as.RawValue),
		"created_at":     as.CreatedAt.AsTime(),
	}
	if len(as.Value) > 0 {
		ar["value"] = as.Value
	}

	if exists {

		or := goqu.Record{
			"key":            os.Key,
			"channel_id":     os.ChannelId,
			"transaction_id": os.TransactionId,
			"type":           os.Type,
			"raw_value":      hex.EncodeToString(os.RawValue),
			"created_at":     os.CreatedAt.AsTime(),
		}
		if len(as.Value) > 0 {
			or["value"] = os.Value
		}

		_, err = txx.
			Insert(oldState).
			Rows(or).Executor().Exec()
		if err != nil {
			return fmt.Errorf("insert old state: %w", err)
		}
		_, err = txx.Update(state).
			Where(goqu.Ex{"key": as.Key}).
			Set(ar).
			Executor().Exec()
		if err != nil {
			return fmt.Errorf("update actual state: %w", err)
		}
	} else {
		_, err := txx.
			Insert(state).
			Rows(ar).Executor().Exec()
		if err != nil {
			return fmt.Errorf("insert actual state: %w", err)
		}
	}

	return nil
}
