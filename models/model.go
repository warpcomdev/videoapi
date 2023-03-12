package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/warpcomdev/videoapi/store"
)

// Model is the base for all models.
// Oracle always caps lock...
type Model struct {
	ID         string    `json:"id" db:"ID"`
	CreatedAt  time.Time `json:"created_at" db:"CREATED_AT"`
	ModifiedAt time.Time `json:"modified_at" db:"MODIFIED_AT"`
}

// Descriptor describes a table
type Descriptor struct {
	TableName string
	FilterSet store.FilterSet
	Create    string
}

// GetID returns video ID
func (m Model) GetID() string {
	return m.ID
}

// Prepare a model to be created into the database
// Return list of columns to be written
func (m *Model) PrepareCreate() ([]string, error) {
	if m.ID == "" {
		return nil, errors.New("missing mandatory attribute id")
	}
	m.CreatedAt = time.Now()
	m.ModifiedAt = m.CreatedAt
	return []string{"ID", "CREATED_AT", "MODIFIED_AT"}, nil
}

// Prepare a model to be updated to the database
// Return list of columns to be updated
func (m *Model) PrepareUpdate(id string) ([]string, error) {
	if id == "" {
		return nil, errors.New("missing mandatory attribute id")
	}
	m.ID = id
	m.ModifiedAt = time.Now()
	return []string{"MODIFIED_AT"}, nil
}

func (d Descriptor) CreateDb(ctx context.Context, db *sqlx.DB) error {
	var sb strings.Builder
	sb.WriteString("CREATE TABLE ")
	sb.WriteString(d.TableName)
	sb.WriteString(" ")
	sb.WriteString(d.Create)
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(sb.String()); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
