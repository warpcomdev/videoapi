package policy

import (
	"context"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
)

// UswrPolicy implements store.Resource and enforces policy on user updates
type UserPolicy struct {
	UserStore store.Resource[models.User]
}

// GetById only allowed to ROLE_ADMIN. Other users can only get themselves.
func (up UserPolicy) GetById(ctx context.Context, id string) (zero models.User, err error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return zero, err
	}
	if claims.Role != models.ROLE_ADMIN {
		if claims.Subject != id {
			return zero, crud.ErrUnauthorized
		}
	}
	return up.UserStore.GetById(ctx, id)
}

// Get denied except to ROLE_ADMIN
func (up UserPolicy) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) ([]models.User, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != models.ROLE_ADMIN {
		return nil, crud.ErrUnauthorized
	}
	return up.UserStore.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post only allowed to admin role
func (up UserPolicy) Post(ctx context.Context, data models.User) (string, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return "", err
	}
	if claims.Role != models.ROLE_ADMIN {
		return "", crud.ErrUnauthorized
	}
	return up.UserStore.Post(ctx, data)
}

// Put restricted depending on role
func (up UserPolicy) Put(ctx context.Context, id string, data models.User) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.ID == id {
		// No user can change its own role
		if data.Role != models.ROLE_UNSET {
			return crud.ErrUnauthorized
		}
	}
	if claims.Role != models.ROLE_ADMIN {
		// Only admin can change other users
		if claims.Subject != id {
			return crud.ErrUnauthorized
		}
		// Only admin can change user roles
		if data.Role != models.ROLE_UNSET {
			return crud.ErrUnauthorized
		}
	}
	return up.UserStore.Put(ctx, id, data)
}

// Delete only allowed to admin role
func (up UserPolicy) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.ID == id {
		// No user can delete himself
		return crud.ErrUnauthorized
	}
	// Only admin can delete users
	if claims.Role != models.ROLE_ADMIN {
		return crud.ErrUnauthorized
	}
	return up.UserStore.Delete(ctx, id)
}
