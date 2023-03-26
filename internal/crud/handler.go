package crud

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Frontend implements a CRUD frontend service
type Frontend interface {
	Get(r *http.Request) (io.ReadCloser, error)
	Post(r *http.Request) (io.ReadCloser, error)
	Put(r *http.Request) error
	Delete(r *http.Request) error
}

func NewHandler(crud Frontend) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var (
			result  io.ReadCloser
			emptyOk bool
			err     error
		)
		defer exhaust(r.Body)
		switch r.Method {
		case http.MethodGet:
			result, err = crud.Get(r)
		case http.MethodPost:
			result, err = crud.Post(r)
			emptyOk = true
		case http.MethodPut:
			err = crud.Put(r)
			emptyOk = true
		case http.MethodDelete:
			err = crud.Delete(r)
			emptyOk = true
		default:
			err = ErrUnsupportedMethod
		}
		query := r.URL.Query()
		redirectOnError := query.Get("redirectOnError")
		redirectOnSuccess := query.Get("redirectOnSuccess")
		// If err != nil and redirectOnError is set, try to redirect to the error URL
		// we append the error message as a query parameter
		if err != nil {
			if redirectOnError != "" {
				errorURL, parseErr := url.Parse(redirectOnError)
				if parseErr == nil {
					errorURL.Query().Set("error", err.Error())
					redirectOnError = errorURL.String()
				}
				http.Redirect(w, r, redirectOnError, http.StatusTemporaryRedirect)
				return
			}
		} else {
			if redirectOnSuccess != "" {
				http.Redirect(w, r, redirectOnSuccess, http.StatusTemporaryRedirect)
				return
			}
		}
		// If no redirecting, just write the reply
		jsonReply(result, err, w, emptyOk)
	}
	return http.HandlerFunc(handler)
}

type queryError struct {
	Error string `json:"error"`
}

// JsonError writes an error to the response
func JsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	reply := queryError{Error: err.Error()}
	var knownError Error
	if errors.As(err, &knownError) {
		code, msg := knownError.HttpError()
		reply.Error = msg
		w.WriteHeader(code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(reply)
}

func jsonReply(resp io.ReadCloser, err error, w http.ResponseWriter, emptyOk bool) {
	if err != nil {
		JsonError(w, err)
		return
	}
	if resp == nil {
		if emptyOk {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		JsonError(w, ErrNotFound)
		return
	}
	defer resp.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("Failed to deliver GET reply: %s", err.Error())
	}
}
