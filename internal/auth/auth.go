package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/warpcomdev/videoapi/internal/crud"
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
	CookieName string    = "VIDEOAPI_SESSION"
)

// auth returns the role of the user in the request
func auth(r *http.Request, jwtKey []byte) (Claims, error) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Authorization header has precedence over cookie
		parts := strings.Split(auth, " ")
		if len(parts) != 2 {
			return Claims{}, crud.ErrorInvalidAuthHeader
		}
		if strings.ToLower(parts[0]) != "bearer" {
			return Claims{}, crud.ErrorInvalidAuthHeader
		}
		auth = parts[1]
	} else {
		// Cookie is supported for posting uploads in a form
		authCookie, err := r.Cookie(CookieName)
		if err != nil {
			return Claims{}, crud.ErrorMisingAuthHeader
		}
		auth = authCookie.Value
	}
	if auth == "" {
		return Claims{}, crud.ErrorMisingAuthHeader
	}
	var currClaims Claims
	token, err := jwt.ParseWithClaims(auth, &currClaims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if token.Method != signingMethod {
			return nil, crud.ErrorUnexpectedSigningMethod
		}
		return jwtKey, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, crud.ErrorInvalidToken
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
		return Claims{}, crud.ErrorMissingRole
	}
	return claims, nil
}

// WithClaims appends Role information to the request context
func WithClaims(jwtKey []byte, handler http.Handler) http.Handler {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		role, err := auth(r, jwtKey)
		if err != nil {
			WriteError(w, err, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsID, role)
		handler.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(wrapper)
}

type errResult struct {
	Error string `json:"error"`
}

func WriteError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(errResult{Error: err.Error()})
}
