package validation

func ValidatePassword(password string) (string, error) {
	if len(password) < 8 || len(password) > 72 {
		return "", ErrInvalidPassword
	}

	return password, nil
}
