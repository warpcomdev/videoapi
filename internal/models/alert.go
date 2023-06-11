package models

import (
	"errors"
	"time"

	"github.com/warpcomdev/videoapi/internal/store"
)

type Alert struct {
	Model
	Name           string    `json:"name" db:"NAME"`
	Timestamp      time.Time `json:"timestamp" db:"TIMESTAMP"`
	Camera         string    `json:"camera" db:"CAMERA"`
	Severity       string    `json:"severity" db:"SEVERITY"`
	Message        string    `json:"message" db:"MESSAGE"`
	AcknowledgedAt NullTime  `json:"acknowledged_at,omitempty" db:"ACKNOWLEDGED_AT"`
	ResolvedAt     NullTime  `json:"resolved_at,omitempty" db:"RESOLVED_AT"`
}

// PrepareCreate prepares an Alert object for persistence
// Returns list of fields to save
func (v *Alert) PrepareCreate() ([]string, error) {
	if v.Name == "" {
		v.Name = v.GetID()[:128]
	}
	if v.Camera == "" {
		return nil, errors.New("missing mandatory attribute camera")
	}
	if v.Timestamp.IsZero() {
		return nil, errors.New("missing mandatory attribute timestamp")
	}
	if v.Severity == "" {
		return nil, errors.New("missing mandatory attribute severity")
	}
	if v.Message == "" {
		return nil, errors.New("missing mandatory attribute message")
	}
	v.Message = v.Message[:512]
	cols, err := v.Model.PrepareCreate()
	if err != nil {
		return nil, err
	}
	cols = append(cols, "NAME", "TIMESTAMP", "CAMERA", "SEVERITY", "MESSAGE")
	return cols, nil
}

// PrepareUpdate prepares an Alert object for update
// Returns list of fileds to update
func (v *Alert) PrepareUpdate(id string) ([]string, error) {
	cols, err := v.Model.PrepareUpdate(id)
	if err != nil {
		return nil, err
	}
	if v.AcknowledgedAt.Valid {
		cols = append(cols, "ACKNOWLEDGED_AT")
	}
	if v.ResolvedAt.Valid {
		cols = append(cols, "RESOLVED_AT")
	}
	return cols, nil
}

// VideoDescriptor describes the Video table (returns name and filterset)
func AlertDescriptor() Descriptor {
	return Descriptor{
		TableName: "ALERTS",
		FilterSet: store.FilterSet{
			"name":            store.StringDbType{},
			"created_at":      store.TimeDbType{},
			"modified_at":     store.TimeDbType{},
			"timestamp":       store.TimeDbType{},
			"severity":        store.StringDbType{},
			"acknowledged_at": store.TimeDbType{},
			"resolved_at":     store.TimeDbType{},
		},
		Create: `
		(
			ID VARCHAR2(256) NOT NULL PRIMARY KEY,
			CREATED_AT TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			MODIFIED_AT TIMESTAMP(6) WITH TIME ZONE,
			NAME VARCHAR(128) NOT NULL,
			TIMESTAMP TIMESTAMP(6) WITH TIME ZONE NOT NULL,
			CAMERA VARCHAR2(128) NOT NULL,
			SEVERITY VARCHAR2(32) NOT NULL,
			MESSAGE VARCHAR2(512) NOT NULL,
			ACKNOWLEDGED_AT TIMESTAMP(6) WITH TIME ZONE NULL,
			RESOLVED_AT TIMESTAMP(6) WITH TIME ZONE NULL,
			CONSTRAINT FK_ALERT_CAMERA FOREIGN KEY (CAMERA) REFERENCES CAMERAS(ID)
		)`,
	}
}
