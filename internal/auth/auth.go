package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/warpcomdev/videoapi/internal/models"
)

type Claims struct {
	jwt.RegisteredClaims
	Role models.Role `json:"role"`
	Name string      `json:"name"`
}

var signingMethod = jwt.SigningMethodHS256

type claimsKey int

const claimsID claimsKey = 0

// auth returns the role of the user in the request
func auth(r *http.Request, jwtKey []byte) (models.Role, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return models.ROLE_UNSET, ErrorMisingAuthHeader
	}
	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return models.ROLE_UNSET, ErrorInvalidAuthHeader
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return models.ROLE_UNSET, ErrorInvalidAuthHeader
	}
	var currClaims Claims
	token, err := jwt.ParseWithClaims(parts[1], &currClaims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if token.Method != signingMethod {
			return nil, ErrorUnexpectedSigningMethod
		}
		return jwtKey, nil
	})
	if err != nil {
		return models.ROLE_UNSET, err
	}
	if !token.Valid {
		return models.ROLE_UNSET, ErrorInvalidToken
	}
	switch currClaims.Role {
	case models.ROLE_ADMIN:
		return models.ROLE_ADMIN, nil
	case models.ROLE_READ_WRITE:
		return models.ROLE_READ_WRITE, nil
	case models.ROLE_READ_ONLY:
		return models.ROLE_READ_ONLY, nil
	}
	return models.ROLE_UNSET, nil
}

// ClaimsFrom returns the role of the user in the request
func ClaimsFrom(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(claimsID).(Claims)
	if !ok {
		return Claims{}, ErrorMissingRole
	}
	return claims, nil
}

// WithClaims appends Role information to the request context
func WithClaims(jwtKey []byte, handler http.Handler) http.Handler {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		role, err := auth(r, jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsID, role)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(wrapper)
}
