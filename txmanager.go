package txmanager

import (
	"context"

	"gorm.io/gorm"
)

type ITransactionManager interface {
	SQLTransaction(ctx context.Context, fn func(context.Context) error) error
	GetTx(ctx context.Context) (tx *gorm.DB, ok bool)
}

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) ITransactionManager {
	return &TransactionManager{db: db}
}

type ctxKey string

var CtxKey = ctxKey("github.com/Hidayathamir/txmanager")

func (tm *TransactionManager) SQLTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, isHasExternalTransaction := ctx.Value(CtxKey).(*gorm.DB)

	if !isHasExternalTransaction {
		tx = tm.db.Begin()
		ctx = context.WithValue(ctx, CtxKey, tx)
	}

	err := fn(ctx)

	if !isHasExternalTransaction {
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	}

	return err
}

func (tm *TransactionManager) GetTx(ctx context.Context) (tx *gorm.DB, ok bool) {
	isHasTransaction := ctx.Value(CtxKey) != nil
	if isHasTransaction {
		if tx, ok := ctx.Value(CtxKey).(*gorm.DB); ok {
			return tx, true
		}
	}
	return nil, false
}
