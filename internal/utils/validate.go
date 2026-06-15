package utils

import (
	"errors"
	"strings"
	"unicode"

	"x-smpp-client/internal/message"
)

type SendRequest struct {
	To         string `json:"to"`
	Text       string `json:"text"`
	SourceAddr string `json:"source_addr"`
	Encoding   string `json:"encoding"`
}

func ValidateSendRequest(req *SendRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return errors.New("to is required")
	}

	cleaned := strings.TrimLeft(req.To, "+")
	if len(cleaned) < 4 || len(cleaned) > 15 {
		return errors.New("to must be between 4 and 15 digits")
	}
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return errors.New("to must contain only digits and optional leading +")
		}
	}

	if strings.TrimSpace(req.Text) == "" {
		return errors.New("text is required")
	}
	if len(req.Text) > 10000 {
		return errors.New("text exceeds maximum length of 10000 characters")
	}

	if req.Encoding != "" {
		if _, err := message.ParseEncoding(req.Encoding); err != nil {
			return err
		}
	}

	return nil
}
