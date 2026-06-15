package message

import (
	"fmt"
	"strings"

	"github.com/linxGnu/gosmpp/data"
)

func ParseEncoding(name string) (data.Encoding, error) {
	switch strings.ToLower(name) {
	case "", "gsm", "gsm7":
		return data.GSM7BIT, nil
	case "latin1", "iso-8859-1":
		return data.LATIN1, nil
	case "ucs2", "utf-16", "unicode":
		return data.UCS2, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", name)
	}
}
