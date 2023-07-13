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
		return http.StatusBadRequest, "filter must have the format q-<field>-<op>:value"
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
	case ErrMultipartName:
		return http.StatusBadRequest, "multipart form fields must have name"
	case ErrMultipartNeedsContentType:
		return http.StatusBadRequest, "multipart file field needs content-type header"
	case ErrMultipartTooManyFiles:
		return http.StatusBadRequest, "too many files in multipart form"
	case ErrMultipartNoFile:
		return http.StatusBadRequest, "no file in multipart form"
	case ErrMimeNotSupported:
		return http.StatusBadRequest, "mime type not supported"
	case ErrUnsupportedMediaType:
		return http.StatusUnsupportedMediaType, "unsupported media type"
	case ErrInvalidJson:
		return http.StatusBadRequest, "invalid json format"
	case ErrorMisingAuthHeader:
		return http.StatusUnauthorized, "missing authorization header"
	case ErrorInvalidAuthHeader:
		return http.StatusUnauthorized, "invalid authorization header"
	case ErrorUnexpectedSigningMethod:
		return http.StatusUnauthorized, "unexpected signing method"
	case ErrorInvalidToken:
		return http.StatusUnauthorized, "invalid token"
	case ErrorInvalidRole:
		return http.StatusUnauthorized, "invalid role"
	case ErrorMissingRole:
		return http.StatusUnauthorized, "missing role"
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
	ErrInvalidJson
	ErrMissingResourceId
	ErrUnsupportedMethod
	ErrUnauthorized
	ErrMultipartName
	ErrMultipartNeedsContentType
	ErrMultipartTooManyFiles
	ErrMultipartNoFile
	ErrMimeNotSupported
	ErrUnsupportedMediaType
	ErrorMisingAuthHeader
	ErrorInvalidAuthHeader
	ErrorUnexpectedSigningMethod
	ErrorInvalidToken
	ErrorInvalidRole
	ErrorMissingRole
)
