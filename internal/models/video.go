package models

import (
	"errors"
	"time"

	"github.com/warpcomdev/videoapi/internal/store"
)

type Video struct {
	Model
	Timestamp time.Time  `json:"timestamp" db:"TIMESTAMP"`
	Camera    string     `json:"camera" db:"CAMERA"`
	Tags      JsonList   `json:"tags,omitempty" db:"TAGS"`
	MediaURL  NullString `json:"media_url,omitempty" db:"MEDIA_URL"`
}

// PrepareCreate prepares a Video object for persistence
// Returns list of fields to save
func (v *Video) PrepareCreate() ([]string, error) {
	if v.Camera == "" {
		return nil, errors.New("missing mandatory attribute camera")
	}
	if v.Timestamp.IsZero() {
		return nil, errors.New("missing mandatory attribute timestamp")
	}
	cols, err := v.Model.PrepareCreate()
	if err != nil {
		return nil, err
	}
	cols = append(cols, "TIMESTAMP", "CAMERA")
	if v.Tags.Valid {
		cols = append(cols, "TAGS")
	}
	if v.MediaURL.Valid && v.MediaURL.String != "" {
		cols = append(cols, "MEDIA_URL")
	}
	return cols, nil
}

// PrepareCreate prepares a Video object for update
// Returns list of fileds to update
func (v *Video) PrepareUpdate(id string) ([]string, error) {
	cols, err := v.Model.PrepareUpdate(id)
	if err != nil {
		return nil, err
	}
	if !v.Timestamp.IsZero() {
		cols = append(cols, "TIMESTAMP")
	}
	if v.Tags.Valid {
		cols = append(cols, "TAGS")
	}
	if v.MediaURL.Valid && v.MediaURL.String != "" {
		cols = append(cols, "MEDIA_URL")
	}
	return cols, nil
}

// VideoDescriptor describes the Video table (returns name and filterset)
func VideoDescriptor() Descriptor {
	return Descriptor{
		TableName: "videos",
		FilterSet: store.FilterSet{
			"created_at":  store.TimeDbType{},
			"modified_at": store.TimeDbType{},
			"timestamp":   store.TimeDbType{},
			"camera":      store.StringDbType{},
			"tags":        store.JsonDbType{},
			"mediaURL":    store.StringDbType{},
		},
		Create: `
		(
			ID VARCHAR2(128) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			TIMESTAMP TIMESTAMP(6) WITH TIME ZONE,
			CAMERA VARCHAR2(128),
			TAGS VARCHAR2(256) NULL,
			MEDIA_URL VARCHAR2(256) NULL,
			CONSTRAINT ensure_json CHECK (tags IS JSON)
		)`,
	}
}
