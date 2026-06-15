package repository

import (
	"context"

	"x-smpp-client/internal/models"
)

type Repository interface {
	CreateAccount(ctx context.Context, a *models.Account) error
	GetAccount(ctx context.Context, id string) (*models.Account, error)

	CreateAPIKey(ctx context.Context, k *models.APIKey) error
	GetAPIKey(ctx context.Context, key string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, accountID string) ([]models.APIKey, error)

	CreateLedgerAccount(ctx context.Context, la *models.LedgerAccount) error
	CreateLedgerBalance(ctx context.Context, lb *models.LedgerBalance) error
	GetLedgerAccount(ctx context.Context, accountID string) (*models.LedgerAccount, error)
	GetLedgerBalance(ctx context.Context, accountID string) (*models.LedgerBalance, error)
	ApplyBalanceChange(ctx context.Context, accountID string, change int64, entryType, reference, description string) (*models.LedgerEntry, error)
	GetLedgerEntries(ctx context.Context, ledgerAccountID string, limit, offset int) ([]models.LedgerEntry, error)

	CreateMessage(ctx context.Context, m *models.Message) error
	UpdateMessageStatus(ctx context.Context, id, status string, parts int, cost int64) error
	SaveDeliveryReceipt(ctx context.Context, r *models.DeliveryReceipt) error
}
