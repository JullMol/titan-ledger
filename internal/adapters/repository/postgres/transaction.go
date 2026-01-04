package postgres

import (
	"context"
	"errors"

	"github.com/JullMol/titan-ledger/internal/core/domain"
	"github.com/JullMol/titan-ledger/internal/core/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) ports.TransactionRepository {
	return &TransactionRepository{db: db}
}

// Helper to select between DB Pool (normal) or DB Transaction (when in transaction)
func (r *TransactionRepository) getDB(ctx context.Context) interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
} {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return r.db
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (wallet_id, amount, reference_id, type, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`
	err := r.getDB(ctx).QueryRow(ctx, query, 
		tx.WalletID, 
		tx.Amount, 
		tx.ReferenceID, 
		tx.Type, 
		tx.CreatedAt,
	).Scan(&tx.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		// Handle duplicate reference_id (idempotency)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateTransaction
		}
		return err
	}
	return nil
}