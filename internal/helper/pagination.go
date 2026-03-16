package helper

import (
	"net/http"
	"strconv"
)

// internal/helper/pagination.go
func GetPaginationParams(r *http.Request) (limit, offset, page int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 10 {
		limit = 10
	}

	offset = (page - 1) * limit
	return limit, offset, page
}
