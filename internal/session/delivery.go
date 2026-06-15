package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/linxGnu/gosmpp/pdu"
)

type DeliveryReceipt struct {
	MessageID  string
	Sub        string
	Dlvrd      string
	SubmitDate time.Time
	DoneDate   time.Time
	Status     string
	Err        string
	Text       string
}

func ParseDeliveryReceipt(msg string) (*DeliveryReceipt, error) {
	r := &DeliveryReceipt{}

	parts := strings.Fields(msg)
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(kv[0])
		val := kv[1]

		switch key {
		case "id":
			r.MessageID = val
		case "sub":
			r.Sub = val
		case "dlvrd":
			r.Dlvrd = val
		case "stat":
			r.Status = val
		case "err":
			r.Err = val
		case "text":
			r.Text = val
		case "submit":
			if t, err := time.Parse("0601021504", val); err == nil {
				r.SubmitDate = t
			}
		case "done":
			if t, err := time.Parse("0601021504", val); err == nil {
				r.DoneDate = t
			}
		}
	}

	if r.MessageID == "" {
		return nil, fmt.Errorf("not a valid delivery receipt: missing id field")
	}

	return r, nil
}

func IsDeliveryReceipt(pd *pdu.DeliverSM) bool {
	return pd.EsmClass&0x04 != 0
}
