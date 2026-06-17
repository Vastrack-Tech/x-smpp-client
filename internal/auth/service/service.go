package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"x-smpp-client/infra/cache"
	"x-smpp-client/internal/models"
	"x-smpp-client/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
)

type AccountRepo interface {
	GetAccountByEmail(ctx context.Context, email string) (*models.Account, error)
	GetAccount(ctx context.Context, id string) (*models.Account, error)
}

type AuthService struct {
	repo         AccountRepo
	sessionStore *cache.SessionStore
	jwtSecret    []byte
}

func New(repo AccountRepo, s *cache.SessionStore, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, sessionStore: s, jwtSecret: []byte(jwtSecret)}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *models.Account, error) {
	acc, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := utils.CheckPassword(acc.PasswordHash, password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(acc.ID)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return token, acc, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", ErrTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrTokenExpired
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", ErrTokenExpired
	}

	revoked, err := s.sessionStore.IsTokenRevoked(ctx, sub, tokenStr)
	if err != nil || revoked {
		return "", ErrTokenRevoked
	}

	return sub, nil
}

func (s *AuthService) Logout(ctx context.Context, accountID, tokenStr string) error {
	return s.sessionStore.RevokeToken(ctx, accountID, tokenStr)
}

func (s *AuthService) GetAccount(ctx context.Context, id string) (*models.Account, error) {
	return s.repo.GetAccount(ctx, id)
}

func (s *AuthService) generateToken(accountID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": accountID,
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
