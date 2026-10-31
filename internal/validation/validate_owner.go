package validation

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func ValidateOwner(owner string) (string, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "", ErrInvalidOwner
	}

	ownerLength := utf8.RuneCountInString(owner)
	if ownerLength < 2 || ownerLength > 50 {
		return "", ErrInvalidOwner
	}

	hasLetterOrDigit := false
	for _, char := range owner {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			hasLetterOrDigit = true
		}

		if unicode.IsControl(char) {
			return "", ErrInvalidOwner
		}
	}

	if !hasLetterOrDigit {
		return "", ErrInvalidOwner
	}

	return owner, nil
}
