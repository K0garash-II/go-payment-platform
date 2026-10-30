package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	ErrInvalidJSON  = errors.New("invalid JSON")
	ErrBodyTooLarge = errors.New("request body too large")
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrBodyTooLarge
		}
		return ErrInvalidJSON
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrBodyTooLarge
		}
		return ErrInvalidJSON
	}
	return nil
}
