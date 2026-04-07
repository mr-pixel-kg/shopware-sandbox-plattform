package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
)

func parsePaginationParams(r *http.Request) (limit, offset int, err error) {
	limit = 50
	offset = 0
	q := r.URL.Query()

	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil || parsed <= 0 || parsed > 500 {
			return 0, 0, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "limit must be between 1 and 500"}
		}
		limit = parsed
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		parsed, parseErr := strconv.Atoi(v)
		if parseErr != nil || parsed < 0 {
			return 0, 0, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "offset must be 0 or greater"}
		}
		offset = parsed
	}
	return limit, offset, nil
}

func buildPaginationMeta(count, limit, offset int, total int64) dto.PaginationMeta {
	return dto.PaginationMeta{
		Limit:   limit,
		Offset:  offset,
		Count:   count,
		Total:   total,
		HasMore: int64(offset+count) < total,
	}
}
