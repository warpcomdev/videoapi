package models

import (
	"errors"

	"github.com/warpcomdev/videoapi/internal/store"
)

type Camera struct {
	Model
	Name      string     `json:"name" db:"NAME"`
	Latitude  float64    `json:"latitude" db:"LATITUDE"`
	Longitude float64    `json:"longitude" db:"LONGITUDE"`
	LocalPath NullString `json:"local_path,omitempty" db:"LOCAL_PATH"`
}

// PrepareCreate prepares a Video object for persistence
// Returns list of fields to save
func (v *Camera) PrepareCreate() ([]string, error) {
	if v.Name == "" {
		return nil, errors.New("missing mandatory attribute name")
	}
	if v.Latitude == 0 {
		return nil, errors.New("missing mandatory attribute latitude")
	}
	if v.Longitude == 0 {
		return nil, errors.New("missing mandatory attribute longitude")
	}
	cols, err := v.Model.PrepareCreate()
	if err != nil {
		return nil, err
	}
	cols = append(cols, "NAME", "LATITUDE", "LONGITUDE")
	if v.LocalPath.Valid && v.LocalPath.String != "" {
		cols = append(cols, "LOCAL_PATH")
	}
	return cols, nil
}

// PrepareCreate prepares a Video object for update
// Returns list of fileds to update
func (v *Camera) PrepareUpdate(id string) ([]string, error) {
	cols, err := v.Model.PrepareUpdate(id)
	if err != nil {
		return nil, err
	}
	if v.Name != "" {
		cols = append(cols, "NAME")
	}
	if v.Latitude != 0 {
		cols = append(cols, "LATITUDE")
	}
	if v.Longitude != 0 {
		cols = append(cols, "LONGITUDE")
	}
	if v.LocalPath.Valid && v.LocalPath.String != "" {
		cols = append(cols, "LOCAL_PATH")
	}
	return cols, nil
}

// CameraDescriptor describes the Video table (returns name and filterset)
func CameraDescriptor() Descriptor {
	return Descriptor{
		TableName: "camera",
		FilterSet: store.FilterSet{
			"created_at":  store.TimeDbType{},
			"modified_at": store.TimeDbType{},
			"name":        store.StringDbType{},
		},
		Create: `
		(
			ID VARCHAR2(128) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			NAME VARCHAR2(128) NOT NULL,
			LATITUDE NUMBER(16, 10) NOT NULL,
			LONGITUDE NUMBER(16, 10) NOT NULL,
			LOCAL_PATH VARCHAR2(512) NULL
		)`,
	}
}
