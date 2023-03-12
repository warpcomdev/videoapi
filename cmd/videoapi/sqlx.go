package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/warpcomdev/videoapi/store"
)

// Builds a "OFFSET x LIMIT x" clause the oracle way
func oracleLimiter(offset, limit int) string {
	return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
}

// Basic implementation of Querier for sqlx
// TODO: add statement cache
type SqlxQuerier struct {
	DB *sqlx.DB
}

// GetContext implements store.Querier
func (q SqlxQuerier) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	err := q.DB.GetContext(ctx, dest, replaceOraPlaceholders(query), args...)
	return err
}

// SelectContext implements store.Querier
func (q SqlxQuerier) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	err := q.DB.SelectContext(ctx, dest, replaceOraPlaceholders(query), args...)
	return err
}

// Basic implementation of Prepared Statement for sqlx
type SqlxStatement struct {
	Stmt        *sql.Stmt
	queryString string
}

// QueryString implements Statement
func (stmt SqlxStatement) QueryString() string {
	return stmt.queryString
}

// Close implements Statement
func (stmt SqlxStatement) Close() error {
	return stmt.Stmt.Close()
}

// Execute implements Statement
func (stmt SqlxStatement) Execute(ctx context.Context, params ...any) error {
	_, err := stmt.Stmt.ExecContext(ctx, params...)
	return err
}

// Basic implementations of Transaction for sqlx
// TODO: Add statement cache
type SqlxTransaction struct {
	Tx *sqlx.Tx
}

// PrepareNamed implements Transaction
func (tx SqlxTransaction) PrepareNamed(ctx context.Context, sql string, params interface{}) (store.Statement, []any, error) {
	sql, args, err := tx.Tx.BindNamed(sql, params)
	if err != nil {
		return nil, nil, err
	}
	// For some reason, go-ora library does not provide the proper placeholders
	// to sqlx. So we must do our own replacement
	sql = replaceOraPlaceholders(sql)
	stmt, err := tx.Tx.PrepareContext(ctx, sql)
	if err != nil {
		return nil, nil, err
	}
	return SqlxStatement{Stmt: stmt, queryString: sql}, args, nil
}

func replaceOraPlaceholders(query string) string {
	// for some reason, go-ora does not seem to replace placeholders properly...
	// Replace every '?' by consecutive :1, :2, etc
	var sb strings.Builder
	match := 1
	for {
		index := strings.Index(query, "?")
		if index < 0 {
			sb.WriteString(query)
			return sb.String()
		}
		sb.WriteString(query[0:index])
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(match))
		match += 1
		if len(query) <= index+1 {
			return sb.String()
		}
		query = query[index+1:]
	}
}

// Commit implements store.Transaction
func (tx SqlxTransaction) Commit() error {
	return tx.Tx.Commit()
}

// Rollback implements store.Transaction
func (tx SqlxTransaction) Rollback() {
	tx.Tx.Rollback()
}

// Basic implementations of store.Executor for sqlx
type SqlxExecutor struct {
	DB *sqlx.DB
}

// Begin implements store.Executor
func (e SqlxExecutor) Begin(ctx context.Context) (store.Transaction, error) {
	tx, err := e.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return SqlxTransaction{Tx: tx}, nil
}
