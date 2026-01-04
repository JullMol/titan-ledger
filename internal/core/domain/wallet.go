package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidAmount 	    = errors.New("amount must be greater than zero")
	ErrDuplicateTransaction = errors.New("transaction already processed")
)

type Wallet struct {
	ID			string
	UserID		string
	Balance		int64
	Currency 	string
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
}

func NewWallet(userID string, currency string) *Wallet {
	return &Wallet{
		UserID:   userID,
		Balance:  0,
		Currency: currency,
	}
}

func (w *Wallet) HasSufficientBalance(amount int64) bool {
	return w.Balance >= amount
}