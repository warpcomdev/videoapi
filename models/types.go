package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Particular type of string that contains a json array of tags
type JsonList struct {
	sql.NullString
	Populated bool
}

type NullString struct {
	sql.NullString
	Populated bool
}

type NullTime struct {
	sql.NullTime
	Populated bool
}

// Scan the field as a json array
func (n JsonList) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("[]"), nil
	}
	return []byte(n.String), nil
}

// Value turns the array into a database string
func (n *JsonList) UnmarshalJSON(data []byte) error {
	if data == nil {
		return errors.New("field should be optional")
	}
	var valid []string
	if err := json.Unmarshal(data, &valid); err != nil {
		return err
	}
	text, err := json.Marshal(valid)
	if err != nil {
		return err
	}
	n.Populated = true
	n.Valid = true
	n.String = string(text)
	return nil
}

// Scan the field as a json array
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

// Value turns the array into a database string
func (n *NullString) UnmarshalJSON(data []byte) error {
	if data == nil {
		return errors.New("field should be optional")
	}
	var valid string
	if err := json.Unmarshal(data, &valid); err != nil {
		return err
	}
	n.Populated = true
	n.Valid = true
	n.String = valid
	return nil
}

// Scan the field as a json array
func (n NullTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}

// Value turns the array into a database string
func (n *NullTime) UnmarshalJSON(data []byte) error {
	if data == nil {
		return errors.New("field should be optional")
	}
	var valid time.Time
	if err := json.Unmarshal(data, &valid); err != nil {
		return err
	}
	n.Populated = true
	n.Valid = true
	n.Time = valid
	return nil
}
