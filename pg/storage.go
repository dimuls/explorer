package pg

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"explorer/ent"
	_ "explorer/pg/migrations"
)

type Storage struct {
	dsn   string
	sqlDB *sql.DB
	db    *goqu.Database
	log   *logrus.Entry
}

func NewStorage(dsn string) (*Storage, error) {
	sqlDB, err := sql.Open(postgresDialect, dsn)
	if err != nil {
		return nil, err
	}

	db := goqu.New(postgresDialect, sqlDB)

	return &Storage{
		dsn:   dsn,
		sqlDB: sqlDB,
		db:    db,
		log: logrus.WithFields(logrus.Fields{
			"subsystem": "pg_storage",
		}),
	}, nil
}

func (s *Storage) Migrate() error {

	sqlDB, err := sql.Open(postgresDialect, s.dsn)
	if err != nil {
		return fmt.Errorf("open DB: %w", err)
	}

	d, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"embed://", postgresDialect, d)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.sqlDB.Close()
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

func (s *Storage) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.Db.BeginTx(ctx, nil)
}

func (s *Storage) AddPeerTx(ctx context.Context, tx *sql.Tx,
	p ent.Peer) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				s.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	var pp ent.Peer

	exists, err := txx.Select().
		From(peer).
		Where(goqu.Ex{"url": p.URL}).
		ScanStructContext(ctx, &pp)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return pp.ID, nil
	}

	_, err = txx.
		Insert(peer).
		Rows(p).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add peer to DB: %w", err)
	}
	return
}

func (s *Storage) AddChannelTx(ctx context.Context, tx *sql.Tx,
	c ent.Channel) error {
	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(channel).
		Rows(c).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (s *Storage) AddChannelConfigTx(ctx context.Context, tx *sql.Tx,
	cc ent.ChannelConfig) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(channelConfig).
		Rows(goqu.Record{
			"channel_id": cc.ChannelID,
			"raw":        hex.EncodeToString(cc.Raw),
			"parsed":     cc.Parsed,
			"created_at": cc.CreatedAt}).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (s *Storage) AddPeerChannelTx(ctx context.Context, tx *sql.Tx,
	c ent.PeerChannel) error {
	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(peerChannel).
		Rows(c).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (s *Storage) AddChaincodeTx(ctx context.Context, tx *sql.Tx,
	c ent.Chaincode) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				s.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	var cc ent.Chaincode

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.
		Select().
		From(chaincode).
		Where(goqu.Ex{"name": c.Name, "version": c.Version}).
		ScanStructContext(ctx, &cc)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return cc.ID, nil
	}

	_, err = txx.
		Insert(chaincode).
		Rows(c).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add chaincode to DB: %w", err)
	}
	return
}

func (s *Storage) AddChannelChaincodeTx(ctx context.Context, tx *sql.Tx,
	cc ent.ChannelChaincode) error {
	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(channelChaincode).
		Rows(cc).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (s *Storage) AddBlockTx(ctx context.Context, tx *sql.Tx,
	b ent.Block) (id int64, err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				s.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	var bb ent.Block

	txx := goqu.NewTx(postgresDialect, tx)

	exists, err := txx.
		Select().
		From(block).
		Where(goqu.Ex{"channel_id": b.ChannelID, "number": b.Number}).
		ScanStructContext(ctx, &bb)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return bb.ID, nil
	}

	_, err = txx.
		Insert(block).
		Rows(b).
		Returning("id").
		Executor().ScanValContext(ctx, &id)
	if err != nil {
		return 0, fmt.Errorf("add peer to DB: %w", err)
	}
	return
}

func (s *Storage) AddTransactionTx(ctx context.Context, tx *sql.Tx,
	t ent.Transaction) error {

	_, err := goqu.NewTx(postgresDialect, tx).
		Insert(transaction).
		Cols("id", "block_id", "created_at").
		Rows(goqu.Record{
			"id":         t.ID,
			"block_id":   t.BlockID,
			"created_at": t.CreatedAt,
		}).
		OnConflict(goqu.DoNothing()).
		Executor().ExecContext(ctx)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			s.log.WithError(err2).Error("failed to rollback transaction")
		}
	}
	return err
}

func (s *Storage) AddStateTx(ctx context.Context, tx *sql.Tx,
	as ent.State) (err error) {

	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				s.log.WithError(err2).
					Error("failed to rollback transaction")
			}
		}
	}()

	txx := goqu.NewTx(postgresDialect, tx)

	var os ent.State

	exists, err := txx.Select().
		From(state).
		Where(goqu.Ex{"key": as.Key}).
		Executor().ScanStruct(&os)
	if err != nil {
		return fmt.Errorf("get actual state: %w", err)
	}

	ar := goqu.Record{
		"key":            as.Key,
		"transaction_id": as.TransactionID,
		"raw_value":      hex.EncodeToString(as.RawValue),
	}
	if len(as.Value) > 0 {
		ar["value"] = as.Value
	}

	or := goqu.Record{
		"key":            os.Key,
		"transaction_id": os.TransactionID,
		"raw_value":      hex.EncodeToString(as.RawValue),
	}
	if len(as.Value) > 0 {
		or["value"] = os.Value
	}

	if exists {
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

func (s *Storage) LastBlockID(ctx context.Context) (lastID int64, err error) {
	found, err := s.db.
		From("block").
		Select(goqu.COALESCE(goqu.MAX("id"), 0)).
		Executor().ScanValContext(ctx, &lastID)
	if err != nil {
		return
	}
	if !found {
		lastID = 0
	}
	return
}
