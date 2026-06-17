package models

import "time"

type Account struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type APIKey struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	AccountID string    `json:"account_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type LedgerAccount struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LedgerBalance struct {
	ID              string    `json:"id"`
	LedgerAccountID string    `json:"ledger_account_id"`
	Balance         int64     `json:"balance"`
	Version         int       `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LedgerEntry struct {
	ID              string    `json:"id"`
	LedgerAccountID string    `json:"ledger_account_id"`
	Type            string    `json:"type"`
	Amount          int64     `json:"amount"`
	Reference       string    `json:"reference"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}
