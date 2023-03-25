package store

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/warpcomdev/videoapi/internal/crud"
)

// Resource implements the CRUD operations
type Resource[T any] interface {
	// Get resource by id
	GetById(context.Context, string) (T, error)
	// Get resource by filter
	Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]T, error)
	// Post (create) new resource, return id
	Post(ctx context.Context, data T) (string, error)
	// Put (update) resource
	Put(ctx context.Context, id string, data T) error
	// Delete resource by id
	Delete(context.Context, string) error
}

// Adaptor matches store.Resource to crud.Resource
type Adaptor[T any] struct {
	Resource Resource[T]
}

// Get resource by id
func (vr Adaptor[T]) GetById(ctx context.Context, id string) (io.ReadCloser, error) {
	v, err := vr.Resource.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

type getResult[T any] struct {
	Data []T    `json:"data"`
	Next string `json:"next"`
}

// Get resource list
func (vr Adaptor[T]) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error) {
	vs, err := vr.Resource.Get(ctx, filter, sort, ascending, offset, limit)
	if err != nil {
		return nil, err
	}
	if vs == nil {
		vs = make([]T, 0)
	}
	result := getResult[T]{
		Data: vs,
		Next: "",
	}
	if len(vs) >= limit {
		result.Next = crud.Next(filter, sort, ascending, offset, limit)
	}
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

type postResult struct {
	ID string `json:"id"`
}

// Post (create) new resource
func (vr Adaptor[T]) Post(ctx context.Context, r io.Reader) (io.ReadCloser, error) {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return nil, err
	}
	id, err := vr.Resource.Post(ctx, orig)
	if err != nil {
		return nil, err
	}
	// Return id, in case we generate it in the future
	byteid, err := json.Marshal(postResult{ID: id})
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(byteid)), nil
}

// Put (update) resource
func (vr Adaptor[T]) Put(ctx context.Context, id string, r io.Reader) error {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return err
	}
	if err := vr.Resource.Put(ctx, id, orig); err != nil {
		return err
	}
	return nil
}

// Delete resource by id
func (vr Adaptor[T]) Delete(ctx context.Context, id string) error {
	return vr.Resource.Delete(ctx, id)
}

// Adapt builds a resource for the given model
func Adapt[T any](resource Resource[T]) Adaptor[T] {
	return Adaptor[T]{
		Resource: resource,
	}
}
