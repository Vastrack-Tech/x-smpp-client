package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

type MeResponse struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}
