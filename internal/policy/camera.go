package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// UswrPolicy implements store.Resource and enforces policy on user updates
type CameraPolicy struct {
	CameraStore store.Resource[models.Camera]
}

// GetById allowed to anyone
func (up CameraPolicy) GetById(ctx context.Context, id string) (models.Camera, error) {
	return up.CameraStore.GetById(ctx, id)
}

// Get allowed to anyone
func (up CameraPolicy) Get(ctx context.Context, filter []crud.Filter, outerOp crud.OuterOperation, innerOp crud.InnerOperation, sort []string, ascending bool, offset, limit int) ([]models.Camera, error) {
	return up.CameraStore.Get(ctx, filter, outerOp, innerOp, sort, ascending, offset, limit)
}

// Count allowed to anyone
func (up CameraPolicy) Count(ctx context.Context, filter []crud.Filter, outerOp crud.OuterOperation, innerOp crud.InnerOperation) (uint64, error) {
	return up.CameraStore.Count(ctx, filter, outerOp, innerOp)
}

// Post allowed only to ROLE_ADMIN
func (up CameraPolicy) Post(ctx context.Context, data models.Camera) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN {
		return "", crud.ErrUnauthorized
	}
	return up.CameraStore.Post(ctx, data)
}

// Put restricted depending on role
func (up CameraPolicy) Put(ctx context.Context, id string, data models.Camera) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	if claims.Role != models.ROLE_ADMIN {
		// Read-write users can only change the store path
		if !data.LocalPath.Valid || data.LocalPath.String == "" {
			return crud.ErrUnauthorized
		}
		allowed := models.Camera{
			LocalPath: data.LocalPath,
		}
		data = allowed
	}
	return up.CameraStore.Put(ctx, id, data)
}

// Delete allowed only to ROLE_ADMIN
func (up CameraPolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN {
		return crud.ErrUnauthorized
	}
	return up.CameraStore.Delete(ctx, id)
}
