package crud

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type InnerOperation string
type OuterOperation string

// Operations to merge several separate filters
const (
	OUTER_AND     OuterOperation = "AND"
	OUTER_OR      OuterOperation = "OR"
	OUTER_DEFAULT OuterOperation = OUTER_AND
)

// Operations to merge several values within a single filter
const (
	INNER_AND     InnerOperation = "AND"
	INNER_OR      InnerOperation = "OR"
	INNER_DEFAULT InnerOperation = INNER_AND
)

// Resource implements the CRUD operations
type Resource interface {
	// Get resource by id
	GetById(context.Context, string) (io.ReadCloser, error)
	// Get resource by filter
	Get(ctx context.Context, filter []Filter, outerOp OuterOperation, innerOp InnerOperation, sort []string, ascending bool, offset, limit int, count bool) (io.ReadCloser, error)
	// Post (create) new resource
	Post(context.Context, io.Reader) (io.ReadCloser, error)
	// Put (update) resource
	Put(context.Context, string, io.Reader) error
	// Delete resource by id
	Delete(context.Context, string) error
}

// ResourceFrontend implements Frontend
type ResourceFrontend struct {
	resource Resource
}

// NewResourceHandler for a given kind of Resource
func FromResource(r Resource) ResourceFrontend {
	return ResourceFrontend{
		resource: r,
	}
}

func exhaust(r io.ReadCloser) {
	if r != nil {
		io.Copy(ioutil.Discard, r)
		r.Close()
	}
}

func (h ResourceFrontend) Get(r *http.Request) (io.ReadCloser, error) {
	id := strings.Trim(r.URL.Path, "/")
	if id != "" {
		// Get single entry
		return h.resource.GetById(r.Context(), id)
	}
	// Get paginated entry
	var (
		filter    []Filter
		sort      []string
		ascending bool
		offset    int
		limit     int
		count     bool
		innerOp   InnerOperation
		outerOp   OuterOperation
		err       error
	)
	params := r.URL.Query()
	if asc := params.Get("ascending"); asc != "" {
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
			return nil, err
		}
		offset = intOff
	}
	if lim := params.Get("limit"); lim != "" {
		intLim, err := strconv.Atoi(lim)
		if err != nil {
			return nil, err
		}
		limit = intLim
	}
	if cnt := params.Get("count"); cnt == "true" {
		count = true
	}
	io := strings.ToUpper(params.Get("inner-op"))
	switch InnerOperation(io) {
	case INNER_AND:
		innerOp = INNER_AND
	case INNER_OR:
		innerOp = INNER_OR
	default:
		innerOp = INNER_DEFAULT
	}
	oo := strings.ToUpper(params.Get("outer-op"))
	switch OuterOperation(oo) {
	case OUTER_AND:
		outerOp = OUTER_AND
	case OUTER_OR:
		outerOp = OUTER_OR
	default:
		outerOp = OUTER_DEFAULT
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
					return nil, ErrInvalidColumn
				}
			}
		}
		// Support legacy query parameters format "q:"
		if strings.HasPrefix(k, "q:") {
			other[k] = v
		}
		if strings.HasPrefix(k, "q-") {
			other[k] = v
		}
	}
	filter, err = filtersFrom(other)
	if err != nil {
		return nil, err
	}
	return h.resource.Get(r.Context(), filter, outerOp, innerOp, sort, ascending, offset, limit, count)
}

// Post handler
func (h ResourceFrontend) Post(r *http.Request) (io.ReadCloser, error) {
	if r.Body == nil {
		return nil, ErrEmptyBody
	}
	return h.resource.Post(r.Context(), r.Body)
}

// Put handler
func (h ResourceFrontend) Put(r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return ErrMissingResourceId
	}
	if r.Body == nil {
		return ErrEmptyBody
	}
	return h.resource.Put(r.Context(), id, r.Body)
}

// Delete handler
func (h ResourceFrontend) Delete(r *http.Request) error {
	id := strings.Trim(r.URL.Path, "/")
	if id == "" {
		return ErrMissingResourceId
	}
	return h.resource.Delete(r.Context(), id)
}
