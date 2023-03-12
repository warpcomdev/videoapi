package store

import (
	"fmt"
	"strconv"
	"time"

	"github.com/warpcomdev/videoapi/crud"
)

// DBType represents a database column type that can be filtered
type DbType interface {
	Where(field string, op crud.Operator, val string) (string, interface{}, error)
}

// Set of all fields that can be used as filter
type FilterSet map[string]DbType

// StringDbType represents a string column
type StringDbType struct{}

func sqlOp(op crud.Operator) (string, error) {
	switch op {
	case crud.OP_EQ:
		return "=", nil
	case crud.OP_NE:
		return "<>", nil
	case crud.OP_GT:
		return ">", nil
	case crud.OP_GE:
		return ">=", nil
	case crud.OP_LT:
		return "<", nil
	case crud.OP_LE:
		return "<=", nil
	case crud.OP_LIKE:
		return "like", nil
	default:
		return "", fmt.Errorf("unsupported operator %s", op)
	}
}

func (s StringDbType) Where(field string, op crud.Operator, val string) (string, interface{}, error) {
	textOp, err := sqlOp(op)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%s %s ?", field, textOp), val, nil
}

// IntDbType represents an integer column
type IntDbType struct{}

func (s IntDbType) Where(field string, op crud.Operator, val string) (string, interface{}, error) {
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return "", nil, err
	}
	textOp, err := sqlOp(op)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%s %s ?", field, textOp), intVal, nil
}

// TimeDbType represents a time.Time column
type TimeDbType struct{}

func (s TimeDbType) Where(field string, op crud.Operator, val string) (string, interface{}, error) {
	timeVal, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return "", nil, err
	}
	textOp, err := sqlOp(op)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%s %s ?", field, textOp), timeVal, nil
}

// JsonDbType represents a string column with json format
type JsonDbType struct{}

func (s JsonDbType) Where(field string, op crud.Operator, val string) (string, interface{}, error) {
	switch op {
	case crud.OP_EQ:
		fallthrough
	case crud.OP_LIKE:
		return fmt.Sprintf("%s like ?", field), fmt.Sprintf("%%%s%%", val), nil
	default:
		return "", nil, fmt.Errorf("unsupported operator %s", op)
	}
}
