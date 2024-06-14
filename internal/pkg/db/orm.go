package db

import (
	"context"
	"database/sql"
)

func (s *Storage) Begin(ctx context.Context) (*sql.Tx, error) {
	return s.conn.DB.BeginTx(ctx, nil)
}

func (s *Storage) Get(ctx context.Context, dest any, query string, args ...any) error {
	err := s.conn.DB.GetContext(ctx, dest, query, args...)
	return err
}

func (s *Storage) Insert(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	row := s.conn.DB.QueryRowContext(ctx, query, args...)
	return row, row.Err()
}

func (s *Storage) Update(ctx context.Context, query string, args ...any) error {
	_, err := s.conn.DB.ExecContext(ctx, query, args...)
	return err
}

func (s *Storage) Find(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.conn.DB.QueryContext(ctx, query, args...)
}
