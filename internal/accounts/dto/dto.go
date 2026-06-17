package dto

import "x-smpp-client/internal/models"

type CreatedAccount struct {
	Account *models.Account
	APIKey  *models.APIKey
}

type CreateAccountRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TopUpRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAccountResponse struct {
	Account *AccountInfo `json:"account"`
	APIKey  *APIKeyInfo  `json:"api_key"`
}

type AccountInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type APIKeyInfo struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type BalanceResponse struct {
	AccountID string `json:"account_id"`
	Balance   int64  `json:"balance"`
}
