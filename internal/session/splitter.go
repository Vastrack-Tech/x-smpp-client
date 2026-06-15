package session

import (
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

type SplitResult struct {
	Parts []*pdu.SubmitSM
	Ref   byte
}

func SplitMessage(text string, encoding data.Encoding, srcAddr, destAddr pdu.Address, regDelivery byte) (*SplitResult, error) {
	sm := pdu.NewSubmitSM().(*pdu.SubmitSM)
	sm.SourceAddr = srcAddr
	sm.DestAddr = destAddr
	sm.RegisteredDelivery = regDelivery
	sm.ProtocolID = 0
	sm.ReplaceIfPresentFlag = 0
	sm.EsmClass = 0
	_ = sm.Message.SetMessageWithEncoding(text, encoding)

	parts, err := sm.Split()
	if err != nil {
		return nil, err
	}

	var ref byte
	if len(parts) > 1 {
		if udh := parts[0].Message.UDH(); udh != nil {
			_, _, ref, _ = udh.GetConcatInfo()
		}
	}

	return &SplitResult{Parts: parts, Ref: ref}, nil
}
