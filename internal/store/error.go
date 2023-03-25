package store

import (
	"encoding/json"
	"fmt"
)

// QueryError encapsulates error in oracle Query
type QueryError struct {
	Query   string
	Params  interface{}
	Message string
	Cause   error
}

func (err QueryError) Error() string {
	data, _ := json.Marshal(err.Params)
	if data == nil {
		data = []byte("<marshal failed>")
	}
	return fmt.Sprintf("%s '%s' (%v): %s", err.Message, err.Query, string(data), err.Cause.Error())
}

func (err QueryError) Unwrap() error {
	return err.Cause
}
