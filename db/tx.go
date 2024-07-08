package db

import (
	"context"
	"database/sql"
)

type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) RunInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
