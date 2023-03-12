package crud

import (
	"fmt"
	"net/url"
	"strconv"
)

// Next build Next URL for filtering
func Next(filters []Filter, sort []string, ascending bool, offset, limit int) string {
	query := make(url.Values)
	query.Set("offset", strconv.Itoa(offset+limit))
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
