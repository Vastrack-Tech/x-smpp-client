package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"x-smpp-client/internal/accounts/repository"
	"x-smpp-client/internal/models"
)

const costPerPart int64 = 1

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAccount(ctx context.Context, name, email string) (*models.Account, error) {
	now := time.Now()
	a := &models.Account{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateAccount(ctx, a); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	la := &models.LedgerAccount{
		ID:        uuid.New().String(),
		AccountID: a.ID,
		Currency:  "kobo",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateLedgerAccount(ctx, la); err != nil {
		return nil, fmt.Errorf("create ledger account: %w", err)
	}

	lb := &models.LedgerBalance{
		ID:              uuid.New().String(),
		LedgerAccountID: la.ID,
		Balance:         0,
		Version:         0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateLedgerBalance(ctx, lb); err != nil {
		return nil, fmt.Errorf("create ledger balance: %w", err)
	}

	return a, nil
}

func (s *Service) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	return s.repo.GetAccount(ctx, id)
}

func (s *Service) GetLedgerAccount(ctx context.Context, accountID string) (*models.LedgerAccount, error) {
	return s.repo.GetLedgerAccount(ctx, accountID)
}

func (s *Service) GetBalance(ctx context.Context, accountID string) (int64, error) {
	lb, err := s.repo.GetLedgerBalance(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return lb.Balance, nil
}

func (s *Service) CheckBalance(ctx context.Context, accountID string, estimatedCost int64) error {
	lb, err := s.repo.GetLedgerBalance(ctx, accountID)
	if err != nil {
		return err
	}
	if lb.Balance < estimatedCost {
		return fmt.Errorf("insufficient balance: have %d, need %d", lb.Balance, estimatedCost)
	}
	return nil
}

func (s *Service) TopUpBalance(ctx context.Context, accountID string, amount int64, description string) (*models.LedgerEntry, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	return s.repo.ApplyBalanceChange(ctx, accountID, amount, "credit", "", description)
}

func (s *Service) DeductBalance(ctx context.Context, accountID string, amount int64, reference, description string) (*models.LedgerEntry, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	return s.repo.ApplyBalanceChange(ctx, accountID, -amount, "debit", reference, description)
}

func (s *Service) GetLedgerEntries(ctx context.Context, accountID string, limit, offset int) ([]models.LedgerEntry, error) {
	la, err := s.repo.GetLedgerAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetLedgerEntries(ctx, la.ID, limit, offset)
}

func (s *Service) EstimateCost(parts int) int64 {
	return int64(parts) * costPerPart
}

func (s *Service) ValidateAPIKey(ctx context.Context, key string) (*models.Account, error) {
	k, err := s.repo.GetAPIKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAccount(ctx, k.AccountID)
}

func (s *Service) CreateAPIKey(ctx context.Context, accountID, name string) (*models.APIKey, error) {
	k := &models.APIKey{
		ID:        uuid.New().String(),
		Key:       uuid.New().String(),
		AccountID: accountID,
		Name:      name,
		Active:    true,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateAPIKey(ctx, k); err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return k, nil
}

func (s *Service) CreateMessage(ctx context.Context, m *models.Message) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = m.CreatedAt
	return s.repo.CreateMessage(ctx, m)
}

func (s *Service) UpdateMessageStatus(ctx context.Context, id, status string, parts int, cost int64) error {
	return s.repo.UpdateMessageStatus(ctx, id, status, parts, cost)
}

func (s *Service) SaveDeliveryReceipt(ctx context.Context, r *models.DeliveryReceipt) error {
	return s.repo.SaveDeliveryReceipt(ctx, r)
}
