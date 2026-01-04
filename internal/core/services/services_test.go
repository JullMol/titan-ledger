package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JullMol/titan-ledger/internal/core/domain"
	"github.com/JullMol/titan-ledger/internal/core/ports"
	"github.com/JullMol/titan-ledger/internal/core/services"
)

// Mock repositories
type mockWalletRepo struct {
	wallets  map[string]*domain.Wallet
	saveErr  error
	getErr   error
}

func newMockWalletRepo() *mockWalletRepo {
	return &mockWalletRepo{
		wallets: make(map[string]*domain.Wallet),
	}
}

func (m *mockWalletRepo) Save(ctx context.Context, wallet *domain.Wallet) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	wallet.ID = "generated-id"
	m.wallets[wallet.ID] = wallet
	return nil
}

func (m *mockWalletRepo) GetByID(ctx context.Context, id string) (*domain.Wallet, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	wallet, ok := m.wallets[id]
	if !ok {
		return nil, errors.New("wallet not found")
	}
	return wallet, nil
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, id string, amount int64) error {
	wallet, ok := m.wallets[id]
	if !ok {
		return errors.New("wallet not found")
	}
	wallet.Balance += amount
	return nil
}

type mockTxRepo struct {
	transactions []*domain.Transaction
	createErr    error
}

func (m *mockTxRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	tx.ID = "tx-generated-id"
	m.transactions = append(m.transactions, tx)
	return nil
}

type mockTxManager struct{}

func (m *mockTxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Tests
func TestWalletService_CreateWallet(t *testing.T) {
	repo := newMockWalletRepo()
	service := services.NewWalletService(repo)

	wallet, err := service.CreateWallet(context.Background(), "user123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wallet.UserID != "user123" {
		t.Errorf("expected UserID 'user123', got '%s'", wallet.UserID)
	}
	if wallet.Currency != "IDR" {
		t.Errorf("expected Currency 'IDR', got '%s'", wallet.Currency)
	}
}

func TestWalletService_CreateWallet_EmptyUserID(t *testing.T) {
	repo := newMockWalletRepo()
	service := services.NewWalletService(repo)

	_, err := service.CreateWallet(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty user ID")
	}
}

func TestTransferService_Transfer(t *testing.T) {
	walletRepo := newMockWalletRepo()
	txRepo := &mockTxRepo{}
	txManager := &mockTxManager{}

	// Setup wallets
	walletRepo.wallets["sender"] = &domain.Wallet{ID: "sender", Balance: 100000}
	walletRepo.wallets["receiver"] = &domain.Wallet{ID: "receiver", Balance: 50000}

	service := services.NewTransferService(walletRepo, txRepo, txManager)

	err := service.Transfer(context.Background(), ports.TransferRequest{
		FromWalletID: "sender",
		ToWalletID:   "receiver",
		Amount:       30000,
		ReferenceID:  "txn-001",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify balances
	if walletRepo.wallets["sender"].Balance != 70000 {
		t.Errorf("expected sender balance 70000, got %d", walletRepo.wallets["sender"].Balance)
	}
	if walletRepo.wallets["receiver"].Balance != 80000 {
		t.Errorf("expected receiver balance 80000, got %d", walletRepo.wallets["receiver"].Balance)
	}

	// Verify transaction logs
	if len(txRepo.transactions) != 2 {
		t.Errorf("expected 2 transaction logs, got %d", len(txRepo.transactions))
	}
}

func TestTransferService_InsufficientBalance(t *testing.T) {
	walletRepo := newMockWalletRepo()
	txRepo := &mockTxRepo{}
	txManager := &mockTxManager{}

	// Setup wallets with insufficient balance
	walletRepo.wallets["sender"] = &domain.Wallet{ID: "sender", Balance: 10000}
	walletRepo.wallets["receiver"] = &domain.Wallet{ID: "receiver", Balance: 50000}

	service := services.NewTransferService(walletRepo, txRepo, txManager)

	err := service.Transfer(context.Background(), ports.TransferRequest{
		FromWalletID: "sender",
		ToWalletID:   "receiver",
		Amount:       50000,
		ReferenceID:  "txn-002",
	})

	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestTransferService_InvalidAmount(t *testing.T) {
	walletRepo := newMockWalletRepo()
	txRepo := &mockTxRepo{}
	txManager := &mockTxManager{}

	service := services.NewTransferService(walletRepo, txRepo, txManager)

	err := service.Transfer(context.Background(), ports.TransferRequest{
		FromWalletID: "sender",
		ToWalletID:   "receiver",
		Amount:       0,
		ReferenceID:  "txn-003",
	})

	if err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestTransferService_SelfTransfer(t *testing.T) {
	walletRepo := newMockWalletRepo()
	txRepo := &mockTxRepo{}
	txManager := &mockTxManager{}

	service := services.NewTransferService(walletRepo, txRepo, txManager)

	err := service.Transfer(context.Background(), ports.TransferRequest{
		FromWalletID: "sender",
		ToWalletID:   "sender",
		Amount:       10000,
		ReferenceID:  "txn-004",
	})

	if err == nil {
		t.Error("expected error for self transfer")
	}
}
