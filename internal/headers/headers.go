package headers

import (
	"errors"
	"strings"
	"unicode"
)

type Headers map[string]string

const clrf = "\r\n"

var ERROR_MALFORMED = errors.New("Error headers are malformed")
var ERROR_INVALID_HEADER_FORMAT = errors.New("Error invalid error format")

func NewHeaders() Headers {
	return make(map[string]string)
}

func trimLeftSpace(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	sData := string(data)
	if !strings.Contains(sData, clrf) {
		return 0, false, nil
	} else if sData == clrf {
		return 2, true, nil
	}

	line := strings.Split(sData, clrf)[0]
	bytesConsumed := len(line) + 2

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return 0, false, ERROR_MALFORMED
	}

	headerName := strings.ToLower(trimLeftSpace(parts[0]))
	headerValue := strings.TrimSpace(parts[1])

	if unicode.IsSpace(rune(headerName[len(headerName)-1])) || !validateHeaderName(headerName) {
		return 0, false, ERROR_INVALID_HEADER_FORMAT
	}

	val, ok := h[headerName]
	if ok {
		headerValue = val + ", " + headerValue
	}

	h[headerName] = headerValue

	return bytesConsumed, false, nil
}

func validateHeaderName(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(rune(char)) && !unicode.IsDigit(rune(char)) && !strings.Contains("!#$%&'*+-.^_`|~", string(char)) {
			return false
		}
	}

	return true
}
