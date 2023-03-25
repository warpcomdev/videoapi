package store

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/warpcomdev/videoapi/internal/crud"
)

// Adaptor matches store.Resource to crud.Resource
type Adaptor[T Model, P interface {
	*T
	EditableModel
}] struct {
	Resource Resource[T, P]
	Querier  Querier
	Executor Executor
}

// Get resource by id
func (vr Adaptor[T, P]) GetById(ctx context.Context, id string) (io.ReadCloser, error) {
	v, err := vr.Resource.GetById(ctx, vr.Querier, id)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

type getResult[T Model] struct {
	Data []T    `json:"data"`
	Next string `json:"next"`
}

// Get resource list
func (vr Adaptor[T, P]) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error) {
	vs, err := vr.Resource.Get(ctx, vr.Querier, filter, sort, ascending, offset, limit)
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
func (vr Adaptor[T, P]) Post(ctx context.Context, r io.Reader) (io.ReadCloser, error) {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return nil, err
	}
	id, err := vr.Resource.Post(ctx, vr.Executor, orig)
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
func (vr Adaptor[T, P]) Put(ctx context.Context, id string, r io.Reader) error {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return err
	}
	if err := vr.Resource.Put(ctx, vr.Executor, orig, id); err != nil {
		return err
	}
	return nil
}

// Delete resource by id
func (vr Adaptor[T, P]) Delete(ctx context.Context, id string) error {
	return vr.Resource.Delete(ctx, vr.Executor, id)
}

// Adapt builds a resource for the given model
func Adapt[T Model, P interface {
	*T
	EditableModel
}](tableName string, columns FilterSet, querier Querier, executor Executor, limiter Limiter) Adaptor[T, P] {
	return Adaptor[T, P]{
		Resource: New[T, P](tableName, columns, limiter),
		Querier:  querier,
		Executor: executor,
	}
}
