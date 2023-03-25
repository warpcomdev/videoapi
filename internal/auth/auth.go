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

const (
	claimsID   claimsKey = 0
	cookieName string    = "VIDEOAPI_SESSION"
)

// auth returns the role of the user in the request
func auth(r *http.Request, jwtKey []byte) (Claims, error) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Authorization header has precedence over cookie
		parts := strings.Split(auth, " ")
		if len(parts) != 2 {
			return Claims{}, ErrorInvalidAuthHeader
		}
		if strings.ToLower(parts[0]) != "bearer" {
			return Claims{}, ErrorInvalidAuthHeader
		}
		auth = parts[1]
	} else {
		// Cookie is supported for posting uploads in a form
		authCookie, err := r.Cookie(cookieName)
		if err != nil {
			return Claims{}, ErrorMisingAuthHeader
		}
		auth = authCookie.Value
	}
	if auth == "" {
		return Claims{}, ErrorMisingAuthHeader
	}
	var currClaims Claims
	token, err := jwt.ParseWithClaims(auth, &currClaims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if token.Method != signingMethod {
			return nil, ErrorUnexpectedSigningMethod
		}
		return jwtKey, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, ErrorInvalidToken
	}
	switch currClaims.Role {
	case models.ROLE_ADMIN:
		return currClaims, nil
	case models.ROLE_READ_WRITE:
		return currClaims, nil
	case models.ROLE_READ_ONLY:
		return currClaims, nil
	}
	return Claims{}, nil
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
