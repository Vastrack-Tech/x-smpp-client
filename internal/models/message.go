package models

import "time"

type Message struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	To         string    `json:"to"`
	Text       string    `json:"text"`
	Encoding   string    `json:"encoding"`
	SourceAddr string    `json:"source_addr"`
	Parts      int       `json:"parts"`
	Status     string    `json:"status"`
	Cost       int64     `json:"cost"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DeliveryReceipt struct {
	ID        int64     `json:"id"`
	MessageID string    `json:"message_id"`
	AccountID string    `json:"account_id"`
	Status    string    `json:"status"`
	ErrorCode string    `json:"error_code"`
	Raw       string    `json:"raw"`
	ReceivedAt time.Time `json:"received_at"`
}
