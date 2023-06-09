package models

import (
	"errors"
	"time"

	"github.com/warpcomdev/videoapi/internal/store"
)

type Media struct {
	Model
	Timestamp time.Time  `json:"timestamp" db:"TIMESTAMP"`
	Camera    string     `json:"camera" db:"CAMERA"`
	Tags      JsonList   `json:"tags,omitempty" db:"TAGS"`
	MediaURL  NullString `json:"media_url,omitempty" db:"MEDIA_URL"`
}

// PrepareCreate prepares a Media object for persistence
// Returns list of fields to save
func (v *Media) PrepareCreate() ([]string, error) {
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

// PrepareCreate prepares a Media object for update
// Returns list of fileds to update
func (v *Media) PrepareUpdate(id string) ([]string, error) {
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
		TableName: "VIDEOS",
		FilterSet: store.FilterSet{
			"id":          store.StringDbType{},
			"created_at":  store.TimeDbType{},
			"modified_at": store.TimeDbType{},
			"timestamp":   store.TimeDbType{},
			"camera":      store.StringDbType{},
			"tags":        store.JsonDbType{},
			"media_url":   store.StringDbType{},
		},
		Create: `
		(
			ID VARCHAR2(128) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			TIMESTAMP TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			CAMERA VARCHAR2(128) NOT NULL,
			TAGS VARCHAR2(256) NULL,
			MEDIA_URL VARCHAR2(256) NULL,
			CONSTRAINT VIDEOS_ENSURE_JSON CHECK (TAGS IS JSON),
			CONSTRAINT FK_VIDEO_CAMERA FOREIGN KEY (CAMERA) REFERENCES CAMERAS(ID)
		)`,
	}
}

// PictureDescriptor describes the Picture table (returns name and filterset)
func PictureDescriptor() Descriptor {
	return Descriptor{
		TableName: "PICTURES",
		FilterSet: store.FilterSet{
			"id":          store.StringDbType{},
			"created_at":  store.TimeDbType{},
			"modified_at": store.TimeDbType{},
			"timestamp":   store.TimeDbType{},
			"camera":      store.StringDbType{},
			"tags":        store.JsonDbType{},
			"media_url":   store.StringDbType{},
		},
		Create: `
		(
			ID VARCHAR2(256) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			TIMESTAMP TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			CAMERA VARCHAR2(128) NOT NULL,
			TAGS VARCHAR2(256) NULL,
			MEDIA_URL VARCHAR2(256) NULL,
			CONSTRAINT PICTURES_ENSURE_JSON CHECK (TAGS IS JSON),
			CONSTRAINT FK_PICTURE_CAMERA FOREIGN KEY (CAMERA) REFERENCES CAMERAS(ID)
		)`,
	}
}
