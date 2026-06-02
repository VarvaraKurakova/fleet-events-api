package handlers

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

func getLimitOffset(r *http.Request) (int, int) {
	limit := defaultLimit
	offset := 0

	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	offsetParam := r.URL.Query().Get("offset")
	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}
