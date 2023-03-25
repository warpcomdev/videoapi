package crud

import (
	"strings"
	"unicode"
)

// Filter
type Filter struct {
	Field    string
	Operator Operator
	Values   []string
}

// merge identical strings
func merge(values []string) []string {
	set := make(map[string]struct{})
	for _, v := range values {
		v := strings.TrimSpace(v)
		if v != "" {
			set[v] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

// Check an string is a valid sql column name
func isColumnName(name string) bool {
	for _, char := range name {
		if char > unicode.MaxASCII || (!unicode.IsLetter(char) && char != '_') {
			return false
		}
	}
	return true
}

// FiltersFrom build list of filters from url query values
func filtersFrom(params map[string][]string) ([]Filter, error) {
	filters := make(map[string]Filter, len(params))
	for k, v := range params {
		parts := strings.SplitN(k, ":", 3)
		if len(parts) < 3 {
			return nil, ErrInvalidFilter
		}
		for idx, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				return nil, ErrInvalidFilter
			}
			parts[idx] = part
		}
		f := Filter{
			Field:    parts[1],
			Operator: Operator(parts[2]),
			Values:   v,
		}
		// Avoid SQL injection by checking column name
		if !isColumnName(f.Field) {
			return nil, ErrInvalidColumn
		}
		// Also check the operator is valid
		if !f.Operator.Valid() {
			return nil, ErrInvalidOperator
		}
		normalized := strings.Join([]string{parts[1], parts[2]}, ":")
		existing, ok := filters[normalized]
		if ok {
			existing.Values = append(existing.Values, v...)
		} else {
			existing = f
		}
		filters[normalized] = existing
	}
	result := make([]Filter, 0, len(filters))
	for _, v := range filters {
		v.Values = merge(v.Values)
		result = append(result, v)
	}
	return result, nil
}
