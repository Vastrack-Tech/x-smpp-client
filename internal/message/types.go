package message

import (
	"fmt"
	"strings"

	"github.com/linxGnu/gosmpp/data"
)

type Message struct {
	ID         string
	To         string
	ToTon      byte
	ToNpi      byte
	Text       string
	SourceAddr string
	SourceTon  byte
	SourceNpi  byte
	Encoding   string
}

func ParseEncoding(s string) (data.Encoding, error) {
	switch strings.ToLower(s) {
	case "gsm", "gsm7", "gsm-7":
		return data.GSM7BIT, nil
	case "ucs2", "ucs-2", "unicode":
		return data.UCS2, nil
	case "latin1", "latin-1", "iso-8859-1":
		return data.LATIN1, nil
	case "ascii":
		return data.ASCII, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", s)
	}
}
