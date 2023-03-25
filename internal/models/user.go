package models

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"

	"github.com/warpcomdev/videoapi/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	ROLE_UNSET      Role = ""
	ROLE_READ_ONLY  Role = "READ_ONLY"
	ROLE_READ_WRITE Role = "READ_WRITE"
	ROLE_ADMIN      Role = "ADMIN"
)

type User struct {
	Model
	Name     string `json:"name" db:"NAME"`
	Role     Role   `json:"role" db:"ROLE"`
	Password string `json:"password" db:"HASH"` // hashed before persisting
}

// Scan implements sql.Scanner
func (r *Role) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("invalid type for Role")
	}
	switch str {
	case string(ROLE_READ_ONLY):
		*r = ROLE_READ_ONLY
	case string(ROLE_READ_WRITE):
		*r = ROLE_READ_WRITE
	case string(ROLE_ADMIN):
		*r = ROLE_ADMIN
	default:
		return errors.New("invalid value for Role")
	}
	return nil
}

// Value implements driver.Valuer
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

// PrepareCreate prepares a Video object for persistence
// Returns list of fields to save
func (v *User) PrepareCreate() ([]string, error) {
	if v.Name == "" {
		return nil, errors.New("missing mandatory attribute name")
	}
	if v.Password == "" {
		return nil, errors.New("missing mandatory attribute password")
	}
	cols, err := v.Model.PrepareCreate()
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	v.Password = base64.StdEncoding.EncodeToString(hash)
	switch v.Role {
	case ROLE_READ_ONLY:
	case ROLE_READ_WRITE:
	case ROLE_ADMIN:
	default:
		v.Role = ROLE_READ_ONLY
	}
	cols = append(cols, "NAME", "ROLE", "HASH")
	return cols, nil
}

// PrepareUpdate prepares a Video object for update
// Returns list of fields to update
func (v *User) PrepareUpdate(id string) ([]string, error) {
	cols, err := v.Model.PrepareUpdate(id)
	if err != nil {
		return nil, err
	}
	if v.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		v.Password = base64.StdEncoding.EncodeToString(hash)
		cols = append(cols, "HASH")
	}
	if v.Name != "" {
		cols = append(cols, "NAME")
	}
	if v.Role == ROLE_READ_ONLY || v.Role == ROLE_READ_WRITE || v.Role == ROLE_ADMIN {
		cols = append(cols, "ROLE")
	}
	return cols, nil
}

// UserDescriptor describes the User table (returns name and filterset)
func UserDescriptor() Descriptor {
	return Descriptor{
		TableName: "users",
		FilterSet: store.FilterSet{
			"name": store.StringDbType{},
		},
		Create: `
		(
			ID VARCHAR2(128) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			NAME VARCHAR2(256),
			ROLE VARCHAR2(16),
			HASH VARCHAR2(256)
		)`,
	}
}
