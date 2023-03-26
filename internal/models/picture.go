package models

import "github.com/warpcomdev/videoapi/internal/store"

type Picture Video

// PictureDescriptor describes the Picture table (returns name and filterset)
func PictureDescriptor() Descriptor {
	return Descriptor{
		TableName: "pictures",
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
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			TIMESTAMP TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			CAMERA VARCHAR2(128) NOT NULL,
			TAGS VARCHAR2(256) NULL,
			MEDIA_URL VARCHAR2(256) NULL,
			CONSTRAINT pictures_ensure_json CHECK (tags IS JSON)
		)`,
	}
}
