package ports

import (
	"context"

	"github.com/JullMol/titan-ledger/internal/core/domain"
)

// WalletRepository defines the interface for wallet persistence operations
type WalletRepository interface {
	Save(ctx context.Context, wallet *domain.Wallet) error
	GetByID(ctx context.Context, id string) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, id string, amount int64) error
}

// TransactionRepository defines the interface for transaction persistence operations
type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
}

// DBTransactionManager handles database transactions (BEGIN, COMMIT, ROLLBACK)
type DBTransactionManager interface {
	// RunInTransaction executes the function fn within a database transaction
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}