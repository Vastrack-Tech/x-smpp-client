package dto

type CreateAccountRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TopUpRequest struct {
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type BalanceResponse struct {
	AccountID string `json:"account_id"`
	Balance   int64  `json:"balance"`
}
