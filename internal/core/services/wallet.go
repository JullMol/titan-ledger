package services

import (
	"context"
	"fmt"
	"time"

	"github.com/JullMol/titan-ledger/internal/core/domain"
	"github.com/JullMol/titan-ledger/internal/core/ports"
)

type TitanWalletService struct {
	walletRepo ports.WalletRepository
}

func NewWalletService(w ports.WalletRepository) ports.WalletService {
	return &TitanWalletService{
		walletRepo: w,
	}
}

func (s *TitanWalletService) CreateWallet(ctx context.Context, userID string) (*domain.Wallet, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id cannot be empty")
	}

	newWallet := domain.NewWallet(userID, "IDR")
	newWallet.CreatedAt = time.Now()

	err := s.walletRepo.Save(ctx, newWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return newWallet, nil
}

func (s *TitanWalletService) GetWalletBalance(ctx context.Context, walletID string) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return wallet, nil
}