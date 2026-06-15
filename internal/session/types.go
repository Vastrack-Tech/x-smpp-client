package session

import "github.com/linxGnu/gosmpp/pdu"

type PDUHandler func(pdu.PDU)

type Status int

const (
	StatusDisconnected Status = iota
	StatusConnected
	StatusReconnecting
)
