package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
)

// RbacResource enforces RBAC controls on CRUD operations
type RbacResource struct {
	Resource crud.Resource
}

// GetById implements crud.Resource
func (r RbacResource) GetById(ctx context.Context, id string) (io.ReadCloser, error) {
	return r.Resource.GetById(ctx, id)
}

// Get implements crud.Resource
func (r RbacResource) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error) {
	return r.Resource.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post implements crud.Resource
func (r RbacResource) Post(ctx context.Context, body io.Reader) (io.ReadCloser, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return nil, crud.ErrUnauthorized
	}
	return r.Resource.Post(ctx, body)
}

// Put implements crud.Resource
func (r RbacResource) Put(ctx context.Context, id string, body io.Reader) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	return r.Resource.Put(ctx, id, body)
}

// Delete implements crud.Resource
func (r RbacResource) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN && claims.Role != models.ROLE_READ_WRITE {
		return crud.ErrUnauthorized
	}
	return r.Resource.Delete(ctx, id)
}

// UserRbacResource further restricts write operations to admin only
type UserRbacResource struct {
	Resource crud.Resource
}

// GetById implements crud.Resource
func (r UserRbacResource) GetById(ctx context.Context, id string) (io.ReadCloser, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != models.ROLE_ADMIN {
		if id != claims.Subject {
			return nil, crud.ErrUnauthorized
		}
	}
	return r.Resource.GetById(ctx, id)
}

// Get implements crud.Resource
func (r UserRbacResource) Get(ctx context.Context, filter []crud.Filter, sort []string, ascending bool, offset, limit int) (io.ReadCloser, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != models.ROLE_ADMIN {
		return nil, crud.ErrUnauthorized
	}
	return r.Resource.Get(ctx, filter, sort, ascending, offset, limit)
}

// Post implements crud.Resource
func (r UserRbacResource) Post(ctx context.Context, body io.Reader) (io.ReadCloser, error) {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != models.ROLE_ADMIN {
		return nil, crud.ErrUnauthorized
	}
	return r.Resource.Post(ctx, body)
}

// Put implements crud.Resource
func (r UserRbacResource) Put(ctx context.Context, id string, body io.Reader) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN {
		if id != claims.Subject {
			return crud.ErrUnauthorized
		}
		body, err = removeUserRole(body)
		if err != nil {
			return err
		}
	} else {
		// Admins can't change their own role
		if id == claims.Subject {
			body, err = removeUserRole(body)
			if err != nil {
				return err
			}
		}
	}
	return r.Resource.Put(ctx, id, body)
}

// remove "role" field from update body
func removeUserRole(body io.Reader) (io.Reader, error) {
	var user models.User
	if err := json.NewDecoder(body).Decode(&user); err != nil {
		return nil, err
	}
	user.Role = models.ROLE_UNSET
	restricted, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(restricted), nil
}

// Delete implements crud.Resource
func (r UserRbacResource) Delete(ctx context.Context, id string) error {
	claims, err := auth.ClaimsFrom(ctx)
	if err != nil {
		return err
	}
	if claims.Role != models.ROLE_ADMIN {
		return crud.ErrUnauthorized
	}
	return r.Resource.Delete(ctx, id)
}
