package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// MediaPolicy implements store.Resource and enforces policy on user updates
type MediaPolicy struct {
	MediaStore store.Resource[models.Media]
}

// GetById allowed to anyone
func (up MediaPolicy) GetById(ctx context.Context, id string) (models.Media, error) {
	return up.MediaStore.GetById(ctx, id)
}

// Get allowed to anyone
func (up MediaPolicy) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]models.Media, error) {
	return up.MediaStore.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post denied to READ_OMLY role
func (up MediaPolicy) Post(ctx context.Context, data models.Media) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE && claims.Role != models.ROLE_SERVICE {
		return "", crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.MediaStore.Post(ctx, data)
}

// Put denied to READ_ONLY role
func (up MediaPolicy) Put(ctx context.Context, id string, data models.Media) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE && claims.Role != models.ROLE_SERVICE {
		return crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.MediaStore.Put(ctx, id, data)
}

// Delete denied to READ_ONLY role
func (up MediaPolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	return up.MediaStore.Delete(ctx, id)
}
