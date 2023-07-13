package crud

import (
	"fmt"
	"net/url"
	"strconv"
)

// Next build Next URL for filtering
func Next(filters []Filter, outerOp OuterOperation, innerOp InnerOperation, sort []string, ascending bool, offset, limit int) string {
	return navigate(filters, outerOp, innerOp, sort, ascending, offset+limit, limit)
}

// Next build Next URL for filtering
func Prev(filters []Filter, outerOp OuterOperation, innerOp InnerOperation, sort []string, ascending bool, offset, limit int) string {
	offset -= limit
	if offset < 0 {
		offset = 0
	}
	return navigate(filters, outerOp, innerOp, sort, ascending, offset, limit)
}

// Next build Next URL for filtering
func navigate(filters []Filter, outerOp OuterOperation, innerOp InnerOperation, sort []string, ascending bool, offset, limit int) string {
	query := make(url.Values)
	query.Set("offset", strconv.Itoa(offset))
	query.Set("limit", strconv.Itoa(limit))
	if outerOp != OUTER_DEFAULT {
		query.Set("outerOp", string(outerOp))
	}
	if innerOp != INNER_DEFAULT {
		query.Set("innerOp", string(innerOp))
	}
	if ascending {
		query.Set("asc", "true")
	}
	for _, col := range sort {
		query.Add("sort", col)
	}
	for _, filter := range filters {
		key := fmt.Sprintf("q-%s-%s", filter.Field, filter.Operator)
		for _, val := range filter.Values {
			query.Add(key, val)
		}
	}
	return query.Encode()
}
