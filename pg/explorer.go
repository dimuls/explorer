package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"explorer"
	_ "explorer/pg/migrations"
)

type Explorer struct {
	dsn   string
	sqlDB *sql.DB
	db    *goqu.Database
	log   *logrus.Entry
}

func NewExplorer(dsn string) (*Explorer, error) {
	sqlDB, err := sql.Open(postgresDialect, dsn)
	if err != nil {
		return nil, err
	}

	db := goqu.New(postgresDialect, sqlDB)

	return &Explorer{
		dsn:   dsn,
		sqlDB: sqlDB,
		db:    db,
		log: logrus.WithFields(logrus.Fields{
			"subsystem": "pg_storage",
		}),
	}, nil
}

func (e *Explorer) Migrate() error {

	sqlDB, err := sql.Open(postgresDialect, e.dsn)
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

func (e *Explorer) Close() error {
	return e.sqlDB.Close()
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

	var pp explorer.Peer

	exists, err := txx.Select().
		From(peer).
		Where(goqu.Ex{"url": p.Url}).
		ScanStructContext(ctx, &pp)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			e.log.WithError(err2).Error("failed to rollback transaction")
		}
		return 0, fmt.Errorf("get peer from DB: %w", err)
	}
	if exists {
		return pp.Id, nil
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
