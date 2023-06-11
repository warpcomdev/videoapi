package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

func Handler(apiKey string, alerts store.Resource[models.Alert]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer func() {
				io.Copy(io.Discard, r.Body)
				r.Body.Close()
			}()
		}
		if statusCode, err := handle(w, r, apiKey, alerts); err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func handle(w http.ResponseWriter, r *http.Request, apiKey string, alerts store.Resource[models.Alert]) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}
	if r.Body == nil {
		return http.StatusBadRequest, nil
	}
	bearer := r.Header.Get("Authorization")
	allowed := false
	if bearer == "Bearer "+apiKey {
		allowed = true
	} else {
		if r.URL.Query().Get("apiKey") == apiKey {
			allowed = true
		}
	}
	if !allowed {
		return http.StatusForbidden, nil
	}
	decoder := json.NewDecoder(r.Body)
	var hook template.Data
	if err := decoder.Decode(&hook); err != nil {
		return http.StatusBadRequest, err
	}
	var (
		message   strings.Builder
		alertname string
		camera    string
		severity  string
		separator string
	)
	appendBuilder := func(sb *strings.Builder, items map[string]string) {
		if items == nil {
			return
		}
		for k, v := range items {
			if k == "alertname" {
				alertname = v
				continue
			}
			if k == "camera" {
				camera = v
				continue
			}
			if k == "severity" {
				severity = v
				continue
			}
			sb.WriteString(separator)
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(v)
			separator = ", "
		}
	}
	appendBuilder(&message, hook.CommonLabels)
	appendBuilder(&message, hook.GroupLabels)
	appendBuilder(&message, hook.CommonAnnotations)
	messagePrefix := message.String()
	prevSeparator, prevName, prevCamera, prevSeverity := separator, alertname, camera, severity
	var errList []error
	for _, alert := range hook.Alerts {
		if alert.StartsAt.IsZero() {
			continue
		}
		separator, alertname, camera, severity = prevSeparator, prevName, prevCamera, prevSeverity
		var alertMessage strings.Builder
		alertMessage.WriteString(messagePrefix)
		appendBuilder(&alertMessage, alert.Labels)
		appendBuilder(&alertMessage, alert.Annotations)
		alertModel := models.Alert{
			Model: models.Model{
				ID:         fmt.Sprintf("%s_%s_%s", camera, alertname, alert.StartsAt),
				ModifiedAt: time.Now(),
			},
			Camera:   camera,
			Severity: severity,
			Message:  alertMessage.String(),
		}
		var writeErr error
		if alert.Status == "firing" {
			alertModel.CreatedAt = time.Now()
			// Try to update, in case the alert already existed
			_, writeErr = alerts.Post(r.Context(), alertModel)
		}
		if alert.Status == "resolved" {
			alertModel.ResolvedAt.Valid = true
			alertModel.ResolvedAt.Time = alert.EndsAt
			writeErr = alerts.Put(r.Context(), alertModel.GetID(), alertModel)
		}
		if writeErr != nil {
			errList = append(errList, writeErr)
		}
	}
	if errList != nil {
		log.Printf("some alerts failed to be updated: %s", errors.Join(errList...).Error())
	}
	return http.StatusOK, nil
}
