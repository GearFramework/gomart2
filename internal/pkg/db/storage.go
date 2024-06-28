package db

import (
	"context"
	"sync"
)

type Storage struct {
	sync.RWMutex
	conn *StorageConnection
}

func NewStorage(connectionDSN string) *Storage {
	return &Storage{
		conn: NewConnection(&StorageConnectionConfig{
			ConnectionDSN:   connectionDSN,
			ConnectMaxOpens: 10,
		}),
	}
}

func (s *Storage) Init() error {
	if err := s.conn.Open(); err != nil {
		return err
	}
	_, err := s.conn.DB.ExecContext(context.Background(), `
		CREATE SCHEMA IF NOT EXISTS gomartspace
	`)
	if err != nil {
		return err
	}
	_, err = s.conn.DB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS gomartspace.customers (
			id 			SERIAL NOT NULL PRIMARY KEY,
			login 		VARCHAR(128) UNIQUE,
			password 	VARCHAR(128) NOT NULL,
			balance		DECIMAL(16, 2) NOT NULL DEFAULT 0,
			CONSTRAINT login_idx UNIQUE (login)
		)	
	`)
	if err != nil {
		return err
	}
	_, err = s.conn.DB.ExecContext(context.Background(), `
		ALTER TABLE IF EXISTS gomartspace.customers 
		  ADD COLUMN IF NOT EXISTS withdraw DECIMAL(16, 2) NOT NULL DEFAULT 0
	`)
	if err != nil {
		return err
	}
	_, err = s.conn.DB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS gomartspace.orders (
			number 			VARCHAR(128) PRIMARY KEY,
			customer_id 	INTEGER REFERENCES gomartspace.customers(id)
						  	ON DELETE RESTRICT 
						  	ON UPDATE RESTRICT,
			uploaded_at		pg_catalog.timestamptz DEFAULT now(),
			status			VARCHAR(32) DEFAULT 'NEW',
			accrual 		DECIMAL(16,2) NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}
	_, err = s.conn.DB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS gomartspace.withdrawals (
			number 			VARCHAR(128) PRIMARY KEY,
			customer_id 	INTEGER REFERENCES gomartspace.customers(id)
						  	ON DELETE RESTRICT 
						  	ON UPDATE RESTRICT,
			sum 			DECIMAL(16,2) NOT NULL,
			processed_at	pg_catalog.timestamptz DEFAULT now()
		)
	`)
	return err
}

func (s *Storage) Close() {
	s.conn.Close()
}

func (s *Storage) Ping() error {
	return s.conn.Ping()
}
