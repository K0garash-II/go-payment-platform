package validation

import "errors"

var (
	ErrInvalidPage  = errors.New("invalid page value")
	ErrInvalidLimit = errors.New("invalid limit value")
	ErrPageTooLarge = errors.New("page value is too large")
	ErrInvalidJSON  = errors.New("invalid JSON")
	ErrBodyTooLarge = errors.New("request body too large")
	ErrInvalidOwner = errors.New("invalid owner")
)
