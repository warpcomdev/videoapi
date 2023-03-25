package crud

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Resource interface {
	// Get resource by id
	GetById(context.Context, string) (io.ReadCloser, error)
	// Get resource by filter
	Get(ctx context.Context, filter []Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error)
	// Post (create) new resource
	Post(context.Context, io.Reader) (io.ReadCloser, error)
	// Put (update) resource
	Put(context.Context, string, io.Reader) error
	// Delete resource by id
	Delete(context.Context, string) error
}

// handler implements http.Handler
type handler struct {
	Resource Resource
}

// Handler for a given kind of Resource
func Handler(r Resource) http.Handler {
	return handler{
		Resource: r,
	}
}

func exhaust(r *http.Request) {
	if r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}
}

func jsonReply(resp io.ReadCloser, err error, w http.ResponseWriter, emptyOk bool) error {
	if err != nil {
		return err
	}
	if resp == nil {
		if emptyOk {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return nil
	}
	defer resp.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, resp); err != nil {
		log.Printf("Failed to deliver GET reply: %s", err.Error())
	}
	return nil
}

func (h handler) httpGet(w http.ResponseWriter, r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id != "" {
		// Get single entry
		// HACK
		log.Printf("SEARCHING FOR VIDEO %s\n", id)
		resp, err := h.Resource.GetById(r.Context(), id)
		return jsonReply(resp, err, w, false)
	}
	// Get paginated entry
	var (
		filter    []Filter
		sort      []string
		ascending bool
		offset    int
		limit     int
		err       error
	)
	params := r.URL.Query()
	if asc := params.Get("asc"); asc != "" {
		switch strings.ToLower(asc) {
		case "t":
			ascending = true
		case "true":
			ascending = true
		case "y":
			ascending = true
		case "yes":
			ascending = true
		}
	}
	if off := params.Get("offset"); off != "" {
		intOff, err := strconv.Atoi(off)
		if err != nil {
			return err
		}
		offset = intOff
	}
	if lim := params.Get("limit"); lim != "" {
		intLim, err := strconv.Atoi(lim)
		if err != nil {
			return err
		}
		limit = intLim
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	other := make(map[string][]string)
	for k, v := range params {
		if k == "sort" {
			sort = merge(v)
			for _, s := range sort {
				if !isColumnName(s) {
					return ErrInvalidColumn
				}
			}
		}
		if strings.HasPrefix(k, "q:") {
			other[k] = v
		}
	}
	filter, err = filtersFrom(other)
	if err != nil {
		return err
	}
	resp, err := h.Resource.Get(r.Context(), filter, sort, ascending, offset, limit)
	return jsonReply(resp, err, w, false)
}

// Post handler
func (h handler) httpPost(w http.ResponseWriter, r *http.Request) error {
	if r.Body == nil {
		return ErrEmptyBody
	}
	resp, err := h.Resource.Post(r.Context(), r.Body)
	return jsonReply(resp, err, w, false)
}

// Put handler
func (h handler) httpPut(w http.ResponseWriter, r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return ErrMissingResourceId
	}
	if r.Body == nil {
		return ErrEmptyBody
	}
	err := h.Resource.Put(r.Context(), id, r.Body)
	return jsonReply(nil, err, w, true)
}

// Delete handler
func (h handler) httpDelete(w http.ResponseWriter, r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return ErrMissingResourceId
	}
	err := h.Resource.Delete(r.Context(), id)
	return jsonReply(nil, err, w, true)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer exhaust(r)
	switch r.Method {
	case http.MethodGet:
		err = h.httpGet(w, r)
	case http.MethodPost:
		err = h.httpPost(w, r)
	case http.MethodPut:
		err = h.httpPut(w, r)
	case http.MethodDelete:
		err = h.httpDelete(w, r)
	default:
		err = ErrUnsupportedMethod
	}
	if err != nil {
		var knownError Error
		if errors.As(err, &knownError) {
			code, msg := knownError.HttpError()
			http.Error(w, msg, code)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
