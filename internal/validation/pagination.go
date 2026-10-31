package validation

import (
	"net/http"
	"strconv"
)

func ParsePagination(r *http.Request) (limit, offset int, err error) {
	pageValue := r.URL.Query().Get("page")
	if pageValue == "" {
		pageValue = "1"
	}
	pageInt, err := strconv.Atoi(pageValue)
	if err != nil {
		return 0, 0, ErrInvalidPage
	}
	if pageInt <= 0 {
		return 0, 0, ErrInvalidPage
	}
	limitValue := r.URL.Query().Get("limit")
	if limitValue == "" {
		limitValue = "10"
	}
	limitInt, err := strconv.Atoi(limitValue)
	if err != nil {
		return 0, 0, ErrInvalidLimit
	}
	if limitInt <= 0 || limitInt > 100 {
		return 0, 0, ErrInvalidLimit
	}

	maxInt := int(^uint(0) >> 1)

	if pageInt-1 > maxInt/limitInt {
		return 0, 0, ErrPageTooLarge
	}

	offsetValue := (pageInt - 1) * limitInt

	return limitInt, offsetValue, nil
}
