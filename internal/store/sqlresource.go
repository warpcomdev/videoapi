package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/warpcomdev/videoapi/internal/crud"
)

// Model represents any database table with an ID
type Model interface {
	GetID() string
}

// EditableModel is a pointer to a Model, that allows modifications
type EditableModel interface {
	PrepareCreate() ([]string, error)
	PrepareUpdate(id string) ([]string, error)
}

// SQLResource manages database operations in a givem table
type SQLResource[T Model, P interface {
	*T
	EditableModel
}] struct {
	// Construction params
	querier  Querier
	executor Executor
	// Properties of the Table
	tableName string
	columns   map[string]DbType
	// Properties of the SQL dialect
	limiter Limiter
}

// New creates a Resource for the given table
func New[T Model, P interface {
	*T
	EditableModel
}](querier Querier, exec Executor, tableName string, columns map[string]DbType, limiter Limiter) SQLResource[T, P] {
	return SQLResource[T, P]{
		querier:   querier,
		executor:  exec,
		tableName: tableName,
		columns:   columns,
		limiter:   limiter,
	}
}

// GetById searches the table for the given id
func (r SQLResource[T, P]) GetById(ctx context.Context, id string) (t T, err error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(r.tableName)
	sb.WriteString(" WHERE id=? ")
	sb.WriteString(r.limiter(0, 1))
	if err := r.querier.GetContext(ctx, &t, sb.String(), id); err != nil {
		return t, QueryError{
			Message: "failed to get resource",
			Query:   sb.String(),
			Params:  id,
			Cause:   err,
		}
	}
	return t, nil
}

// Get filtered (and possibly paginated) resources
func (r SQLResource[T, P]) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]T, error) {
	var sb strings.Builder
	pp := make([]interface{}, 0, 16)
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(r.tableName)
	if filter != nil && len(filter) > 0 {
		sb.WriteString(" WHERE (")
		sep := ""
		for _, f := range filter {
			dbtype, ok := r.columns[f.Field]
			if !ok {
				return nil, fmt.Errorf("column %s does not exist", f.Field)
			}
			sb.WriteString(sep)
			sep = ") AND ("
			innerSep := ""
			for _, v := range f.Values {
				var (
					cond string
					val  interface{}
					err  error
				)
				if v == "NULL" {
					switch f.Operator {
					case crud.OP_EQ:
						cond = f.Field + "IS NULL"
					case crud.OP_NE:
						cond = f.Field + "IS NOT NULL"
					default:
						return nil, fmt.Errorf("unsupported operador %s for value NULL", f.Operator)
					}
				} else {
					cond, val, err = dbtype.Where(f.Field, f.Operator, v)
				}
				if err != nil {
					return nil, err
				}
				sb.WriteString(innerSep)
				innerSep = " OR "
				sb.WriteString(cond)
				if val != nil {
					pp = append(pp, val)
				}
			}
		}
		sb.WriteString(")")
	}
	if sort != nil && len(sort) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(strings.Join(sort, ", "))
	} else {
		sb.WriteString(" ORDER BY id")
	}
	if ascending {
		sb.WriteString(" ASC ")
	} else {
		sb.WriteString(" DESC ")
	}
	sb.WriteString(r.limiter(offset, limit))
	var result []T
	if err := r.querier.SelectContext(ctx, &result, sb.String(), pp...); err != nil {
		return nil, QueryError{
			Message: "failed to filter resource",
			Query:   sb.String(),
			Params:  pp,
			Cause:   err,
		}
	}
	return result, nil
}

// Post creates a resource in the database
func (r SQLResource[T, P]) Post(ctx context.Context, t T) (string, error) {
	cols, err := P(&t).PrepareCreate()
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(r.tableName)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(cols, ", "))
	sb.WriteString(") VALUES (:")
	sb.WriteString(strings.ToUpper(strings.Join(cols, ", :")))
	sb.WriteString(")")
	tx, err := r.executor.Begin(ctx)
	if err != nil {
		return "", err
	}
	stmt, args, err := tx.PrepareNamed(ctx, sb.String(), t)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	affected, err := stmt.Execute(ctx, args...)
	if err != nil {
		tx.Rollback()
		return "", QueryError{
			Message: "failed to create resource",
			Query:   stmt.QueryString(),
			Params:  t,
			Cause:   err,
		}
	}
	newID := t.GetID()
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	if affected != 1 {
		return "", fmt.Errorf("expected 1 row affected, got %d", affected)
	}
	return newID, nil
}

// Put updates a resource in the database
func (r SQLResource[T, P]) Put(ctx context.Context, id string, t T) error {
	if id == "" {
		return errors.New("cannot update resource with empty id")
	}
	cols, err := P(&t).PrepareUpdate(id)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	sb.WriteString(r.tableName)
	sb.WriteString(" SET ")
	sep := ""
	for _, col := range cols {
		sb.WriteString(sep)
		sep = ", "
		sb.WriteString(col)
		sb.WriteString("=:")
		sb.WriteString(strings.ToUpper(col))
	}
	sb.WriteString(" WHERE id=:ID")
	tx, err := r.executor.Begin(ctx)
	if err != nil {
		return err
	}
	stmt, args, err := tx.PrepareNamed(ctx, sb.String(), t)
	if err != nil {
		return err
	}
	defer stmt.Close()
	affected, err := stmt.Execute(ctx, args...)
	if err != nil {
		tx.Rollback()
		return QueryError{
			Message: "failed to update resource",
			Query:   stmt.QueryString(),
			Params:  t,
			Cause:   err,
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("resource not found")
	}
	return nil
}

type deleteReq struct {
	ID string `db:"ID"`
}

// Delete a resource from the database
func (r SQLResource[T, P]) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("cannot remove resource with empty id")
	}
	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	sb.WriteString(r.tableName)
	sb.WriteString(" WHERE id=:ID")
	tx, err := r.executor.Begin(ctx)
	if err != nil {
		return err
	}
	stmt, args, err := tx.PrepareNamed(ctx, sb.String(), deleteReq{ID: id})
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Execute(ctx, args...); err != nil {
		tx.Rollback()
		return QueryError{
			Message: "failed to delete resource",
			Query:   stmt.QueryString(),
			Params:  id,
			Cause:   err,
		}
	}
	return tx.Commit()
}
