package crud

import (
	"fmt"
	"net/url"
	"strconv"
)

// Next build Next URL for filtering
func Next(filters []Filter, sort []string, ascending bool, offset, limit int) string {
	return navigate(filters, sort, ascending, offset+limit, limit)
}

// Next build Next URL for filtering
func Prev(filters []Filter, sort []string, ascending bool, offset, limit int) string {
	offset -= limit
	if offset < 0 {
		offset = 0
	}
	return navigate(filters, sort, ascending, offset, limit)
}

// Next build Next URL for filtering
func navigate(filters []Filter, sort []string, ascending bool, offset, limit int) string {
	query := make(url.Values)
	query.Set("offset", strconv.Itoa(offset))
	query.Set("limit", strconv.Itoa(limit))
	if ascending {
		query.Set("asc", "true")
	}
	for _, col := range sort {
		query.Add("sort", col)
	}
	for _, filter := range filters {
		key := fmt.Sprintf("q:%s:%s", filter.Field, filter.Operator)
		for _, val := range filter.Values {
			query.Add(key, val)
		}
	}
	return query.Encode()
}
