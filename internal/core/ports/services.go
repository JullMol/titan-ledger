package ports

import (
	"context"

	"github.com/JullMol/titan-ledger/internal/core/domain"
)

type WalletService interface {
	CreateWallet(ctx context.Context, userID string) (*domain.Wallet, error)
	GetWalletBalance(ctx context.Context, walletID string) (*domain.Wallet, error)
	Deposit(ctx context.Context, walletID string, amount int64) (*domain.Wallet, error)
}

type TransactionService interface {
	Transfer(ctx context.Context, req TransferRequest) error
}

type TransferRequest struct {
	FromWalletID string
	ToWalletID   string
	Amount       int64
	ReferenceID  string 
}