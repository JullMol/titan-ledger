package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JullMol/titan-ledger/internal/core/domain"
	"github.com/JullMol/titan-ledger/internal/core/ports"
)

type TitanTransferService struct {
	walletRepo ports.WalletRepository
	txRepo     ports.TransactionRepository
	dbManager  ports.DBTransactionManager
}

// Constructor
func NewTransferService(w ports.WalletRepository, t ports.TransactionRepository, dbm ports.DBTransactionManager) ports.TransactionService {
	return &TitanTransferService{
		walletRepo: w,
		txRepo:     t,
		dbManager:  dbm,
	}
}

func (s *TitanTransferService) Transfer(ctx context.Context, req ports.TransferRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if req.FromWalletID == req.ToWalletID {
		return errors.New("cannot transfer to self")
	}

	// [ATOMIC] Wrap all database logic in RunInTransaction
	// Note: We use 'txCtx' inside this function, not the initial 'ctx'.
	err := s.dbManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		
		// 1. Check Balance & Validate (Read using txCtx)
		senderWallet, err := s.walletRepo.GetByID(txCtx, req.FromWalletID)
		if err != nil {
			return fmt.Errorf("sender wallet error: %w", err)
		}
		if !senderWallet.HasSufficientBalance(req.Amount) {
			return domain.ErrInsufficientBalance
		}

		// 2. Debit Sender Balance (Write using txCtx)
		err = s.walletRepo.UpdateBalance(txCtx, req.FromWalletID, -req.Amount)
		if err != nil {
			return fmt.Errorf("failed to debit sender: %w", err)
		}

		// 3. Credit Receiver Balance (Write using txCtx)
		err = s.walletRepo.UpdateBalance(txCtx, req.ToWalletID, req.Amount)
		if err != nil {
			return fmt.Errorf("failed to credit receiver: %w", err)
		}

		// 4. Create Transaction Logs (Write using txCtx)
		debitLog := &domain.Transaction{
			WalletID:    req.FromWalletID,
			Amount:      -req.Amount,
			ReferenceID: req.ReferenceID + "-dr",
			Type:        domain.TransactionTypeDebit,
			CreatedAt:   time.Now(),
		}
		if err := s.txRepo.Create(txCtx, debitLog); err != nil {
			return err
		}

		creditLog := &domain.Transaction{
			WalletID:    req.ToWalletID,
			Amount:      req.Amount,
			ReferenceID: req.ReferenceID + "-cr",
			Type:        domain.TransactionTypeCredit,
			CreatedAt:   time.Now(),
		}
		if err := s.txRepo.Create(txCtx, creditLog); err != nil {
			return err
		}

		// If return nil, Transaction Manager will COMMIT.
		// If return error, Transaction Manager will ROLLBACK.
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}