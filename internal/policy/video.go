package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// UswrPolicy implements store.Resource and enforces policy on user updates
type VideoPolicy struct {
	VideoStore store.Resource[models.Video]
}

// GetById allowed to anyone
func (up VideoPolicy) GetById(ctx context.Context, id string) (models.Video, error) {
	return up.VideoStore.GetById(ctx, id)
}

// Get allowed to anyone
func (up VideoPolicy) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]models.Video, error) {
	return up.VideoStore.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post denied to READ_OMLY role
func (up VideoPolicy) Post(ctx context.Context, data models.Video) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return "", crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.VideoStore.Post(ctx, data)
}

// Put denied to READ_ONLY role
func (up VideoPolicy) Put(ctx context.Context, id string, data models.Video) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	// People cannot change the media URL, it will be automatically set by the system
	data.MediaURL.Valid = false
	return up.VideoStore.Put(ctx, id, data)
}

// Delete denied to READ_ONLY role
func (up VideoPolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	return up.VideoStore.Delete(ctx, id)
}
