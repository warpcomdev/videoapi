package crud

import (
	"fmt"
	"net/http"
)

// Static error type
type Error int

// Error implements error
func (err Error) HttpError() (int, string) {
	switch err {
	case ErrInvalidFilter:
		return http.StatusBadRequest, "filter must have the format q:<field>:<op>:value"
	case ErrInvalidColumn:
		return http.StatusBadRequest, "sort or filter by invalid column name"
	case ErrInvalidOperator:
		return http.StatusBadRequest, "filter by invalid operator"
	case ErrNotFound:
		return http.StatusNotFound, "not found"
	case ErrEmptyBody:
		return http.StatusBadRequest, "body must not be empty"
	case ErrMissingResourceId:
		return http.StatusBadRequest, "must provide resource id"
	case ErrUnsupportedMethod:
		return http.StatusMethodNotAllowed, "unsupported method"
	case ErrUnauthorized:
		return http.StatusUnauthorized, "unauthorized"
	default:
		return http.StatusInternalServerError, fmt.Sprintf("error code %d", err)
	}
}

// Error implements error
func (err Error) Error() string {
	_, msg := err.HttpError()
	return msg
}

const (
	NoError Error = iota
	ErrInvalidFilter
	ErrInvalidColumn
	ErrInvalidOperator
	ErrNotFound
	ErrEmptyBody
	ErrMissingResourceId
	ErrUnsupportedMethod
	ErrUnauthorized
)
