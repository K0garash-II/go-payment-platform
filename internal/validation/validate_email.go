package validation

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

func ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)

	if email == "" || utf8.RuneCountInString(email) > 255 {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", ErrInvalidEmail
	}

	return strings.ToLower(email), nil
}
