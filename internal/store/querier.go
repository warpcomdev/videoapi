package store

import (
	"context"
)

// Querier is the interface used by GetById and Get
type Querier interface {
	GetContext(ctx context.Context, result interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, result interface{}, query string, args ...interface{}) error
}

// Transaction is an interface used by Post, Put and Delete
type Transaction interface {
	PrepareNamed(ctx context.Context, sql string, mapping any) (Statement, []any, error)
	Commit() error
	Rollback()
}

type Statement interface {
	QueryString() string
	Execute(ctx context.Context, params ...any) (int, error)
	Close() error
}

// Executor is an interface used by Post, Put and Delete
type Executor interface {
	Begin(ctx context.Context) (Transaction, error)
}

// Limiter builds a "LIMIT X, OFFSET Y" clause
type Limiter func(offset, limit int) string
