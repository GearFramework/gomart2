package db

import (
	"context"
	"github.com/GearFramework/gomart/internal/pkg/alog"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type StorageConnection struct {
	DB        *sqlx.DB
	Config    *StorageConnectionConfig
	pgxConfig *pgx.ConnConfig
	logger    *zap.SugaredLogger
}

func NewConnection(config *StorageConnectionConfig) *StorageConnection {
	return &StorageConnection{
		Config: config,
		logger: alog.NewLogger("info"),
	}
}

func (conn *StorageConnection) Open() error {
	var err error = nil
	if conn.pgxConfig, err = conn.getPgxConfig(); err != nil {
		return err
	}
	return conn.openSqlxViaPooler()
}

// openSqlxViaPooler открытие пула соединений
func (conn *StorageConnection) openSqlxViaPooler() error {
	db := stdlib.OpenDB(*conn.pgxConfig)
	conn.DB = sqlx.NewDb(db, "pgx")
	conn.DB.SetMaxOpenConns(conn.Config.ConnectMaxOpens)
	return nil
}

func (conn *StorageConnection) getPgxConfig() (*pgx.ConnConfig, error) {
	pgxConfig, err := pgx.ParseConfig(conn.Config.ConnectionDSN)
	if err != nil {
		conn.logger.Errorf("Unable to parse DSN: %s", err.Error())
		return nil, err
	}
	return pgxConfig, nil
}

func (conn *StorageConnection) Ping() error {
	return conn.DB.PingContext(context.Background())
}

func (conn *StorageConnection) Close() {
	if conn.Ping() == nil {
		conn.logger.Info("close storage connection")
		if err := conn.DB.Close(); err != nil {
			conn.logger.Error(err.Error())
		}
	}
}
