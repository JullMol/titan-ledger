package domain

import "time"

type TransactionType string

const (
	TransactionTypeCredit TransactionType = "CREDIT"
	TransactionTypeDebit  TransactionType = "DEBIT"
)

type Transaction struct {
	ID			string
	WalletID	string
	Amount		int64
	ReferenceID string
	Type		TransactionType
	CreatedAt	time.Time
}

type Transfer struct {
	ID			string
	FromWalletID string
	ToWalletID   string
	Amount		int64
	Status		string
	CreatedAt	time.Time
}