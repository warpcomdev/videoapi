package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// UswrPolicy implements store.Resource and enforces policy on user updates
type PicturePolicy struct {
	PictureStore store.Resource[models.Picture]
}

// GetById allowed to anyone
func (up PicturePolicy) GetById(ctx context.Context, id string) (models.Picture, error) {
	return up.PictureStore.GetById(ctx, id)
}

// Get allowed to anyone
func (up PicturePolicy) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]models.Picture, error) {
	return up.PictureStore.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post denied to READ_OMLY role
func (up PicturePolicy) Post(ctx context.Context, data models.Picture) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return "", crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.PictureStore.Post(ctx, data)
}

// Put denied to READ_ONLY role
func (up PicturePolicy) Put(ctx context.Context, id string, data models.Picture) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.PictureStore.Put(ctx, id, data)
}

// Delete denied to READ_ONLY role
func (up PicturePolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	return up.PictureStore.Delete(ctx, id)
}
