package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// UswrPolicy implements store.Resource and enforces policy on user updates
type AlertPolicy struct {
	AlertStore store.Resource[models.Alert]
}

// GetById allowed to anyone
func (up AlertPolicy) GetById(ctx context.Context, id string) (models.Alert, error) {
	return up.AlertStore.GetById(ctx, id)
}

// Get allowed to anyone
func (up AlertPolicy) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]models.Alert, error) {
	return up.AlertStore.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post allowed to anyone with write permissions
func (up AlertPolicy) Post(ctx context.Context, data models.Alert) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return "", crud.ErrUnauthorized
	}
	return up.AlertStore.Post(ctx, data)
}

// Put only allowed for acknowledge or resolve
func (up AlertPolicy) Put(ctx context.Context, id string, data models.Alert) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	if claims.Role != models.ROLE_ADMIN {
		// Read-write users can only change the store path
		allowed := models.Alert{
			AcknowledgedAt: data.AcknowledgedAt,
			ResolvedAt:     data.ResolvedAt,
		}
		data = allowed
	}
	return up.AlertStore.Put(ctx, id, data)
}

// Delete allowed only to ROLE_ADMIN
func (up AlertPolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN {
		return crud.ErrUnauthorized
	}
	return up.AlertStore.Delete(ctx, id)
}
