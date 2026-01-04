package domain_test

import (
	"testing"

	"github.com/JullMol/titan-ledger/internal/core/domain"
)

func TestNewWallet(t *testing.T) {
	wallet := domain.NewWallet("user123", "IDR")

	if wallet.UserID != "user123" {
		t.Errorf("expected UserID to be 'user123', got '%s'", wallet.UserID)
	}
	if wallet.Currency != "IDR" {
		t.Errorf("expected Currency to be 'IDR', got '%s'", wallet.Currency)
	}
	if wallet.Balance != 0 {
		t.Errorf("expected Balance to be 0, got %d", wallet.Balance)
	}
}

func TestWallet_HasSufficientBalance(t *testing.T) {
	tests := []struct {
		name     string
		balance  int64
		amount   int64
		expected bool
	}{
		{
			name:     "sufficient balance",
			balance:  100000,
			amount:   50000,
			expected: true,
		},
		{
			name:     "exact balance",
			balance:  50000,
			amount:   50000,
			expected: true,
		},
		{
			name:     "insufficient balance",
			balance:  30000,
			amount:   50000,
			expected: false,
		},
		{
			name:     "zero balance",
			balance:  0,
			amount:   1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := &domain.Wallet{Balance: tt.balance}
			result := wallet.HasSufficientBalance(tt.amount)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDomainErrors(t *testing.T) {
	if domain.ErrInsufficientBalance == nil {
		t.Error("ErrInsufficientBalance should not be nil")
	}
	if domain.ErrInvalidAmount == nil {
		t.Error("ErrInvalidAmount should not be nil")
	}
	if domain.ErrDuplicateTransaction == nil {
		t.Error("ErrDuplicateTransaction should not be nil")
	}
}
