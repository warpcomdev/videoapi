package store

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/warpcomdev/videoapi/internal/crud"
)

// adaptor matches store.Resource to crud.Resource
type adaptor[T Model, P interface {
	*T
	EditableModel
}] struct {
	builder  Resource[T, P]
	querier  Querier
	executor Executor
}

// Get resource by id
func (vr adaptor[T, P]) GetById(ctx context.Context, id string) (io.ReadCloser, error) {
	v, err := vr.builder.GetById(ctx, vr.querier, id)
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
func (vr adaptor[T, P]) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error) {
	vs, err := vr.builder.Get(ctx, vr.querier, filter, sort, ascending, offset, limit)
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
func (vr adaptor[T, P]) Post(ctx context.Context, r io.Reader) (io.ReadCloser, error) {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return nil, err
	}
	id, err := vr.builder.Post(ctx, vr.executor, orig)
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
func (vr adaptor[T, P]) Put(ctx context.Context, id string, r io.Reader) error {
	var orig T
	dec := json.NewDecoder(r)
	if err := dec.Decode(&orig); err != nil {
		return err
	}
	if err := vr.builder.Put(ctx, vr.executor, orig, id); err != nil {
		return err
	}
	return nil
}

// Delete resource by id
func (vr adaptor[T, P]) Delete(ctx context.Context, id string) error {
	return vr.builder.Delete(ctx, vr.executor, id)
}

// Adapt builds a resource for the given model
func Adapt[T Model, P interface {
	*T
	EditableModel
}](tableName string, columns FilterSet, querier Querier, executor Executor, limiter Limiter) crud.Resource {
	return adaptor[T, P]{
		builder:  New[T, P](tableName, columns, limiter),
		querier:  querier,
		executor: executor,
	}
}
