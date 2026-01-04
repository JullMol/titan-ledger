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

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) ports.WalletRepository {
	return &WalletRepository{db: db}
}

// Helper to select between DB Pool (normal) or DB Transaction (when in transaction)
func (r *WalletRepository) getDB(ctx context.Context) interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
} {
	// Check if there is a TX in context (injected by TxManager)
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return r.db
}

func (r *WalletRepository) Save(ctx context.Context, wallet *domain.Wallet) error {
	query := `INSERT INTO wallets (user_id, balance, currency) VALUES ($1, $2, $3) RETURNING id, created_at`
	
	// Use r.getDB(ctx) instead of r.db to support transactions
	err := r.getDB(ctx).QueryRow(ctx, query, wallet.UserID, wallet.Balance, wallet.Currency).Scan(&wallet.ID, &wallet.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *WalletRepository) GetByID(ctx context.Context, id string) (*domain.Wallet, error) {
	// Use FOR UPDATE when inside a transaction to prevent race conditions
	// This locks the row until the transaction commits/rollbacks
	query := `SELECT id, user_id, balance, currency, created_at FROM wallets WHERE id = $1 FOR UPDATE`
	
	var w domain.Wallet
	err := r.getDB(ctx).QueryRow(ctx, query, id).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return &w, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, id string, amount int64) error {
	query := `UPDATE wallets SET balance = balance + $1, updated_at = NOW() WHERE id = $2`
	
	tag, err := r.getDB(ctx).Exec(ctx, query, amount, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return domain.ErrInsufficientBalance
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("wallet not found")
	}
	return nil
}