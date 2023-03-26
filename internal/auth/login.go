package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type loginReply struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Token string `json:"token"`
}

type loginConfig struct {
	Secure     bool
	HttpOnly   bool
	SameSite   http.SameSite
	SuperAdmin string
	Expiration time.Duration
	Path       string
}

type AuthOption func(*loginConfig)

// WithSecureCookie changes the default secure cookie flag
func WithSecureCookie(secure bool) AuthOption {
	return func(config *loginConfig) {
		config.Secure = secure
	}
}

// WithHttpOnlyCookie changes the default httpOnly cookie flag
func WithHttpOnlyCookie(httpOnly bool) AuthOption {
	return func(config *loginConfig) {
		config.HttpOnly = httpOnly
	}
}

// WithSameSiteCookie changes the default sameSite cookie flag
func WithSameSiteCookie(sameSite bool) AuthOption {
	return func(config *loginConfig) {
		if sameSite {
			config.SameSite = http.SameSiteStrictMode
		} else {
			config.SameSite = http.SameSiteLaxMode
		}
	}
}

// WithSuperAdmin sets the super admin password
func WithSuperAdmin(password string) AuthOption {
	return func(config *loginConfig) {
		config.SuperAdmin = password
	}
}

// Changes the default session duration
func WithDuration(duration time.Duration) AuthOption {
	return func(config *loginConfig) {
		config.Expiration = duration
	}
}

// Changes the default cookie path
func WithCookiePath(path string) AuthOption {
	return func(config *loginConfig) {
		config.Path = path
	}
}

func applyOptions(options ...AuthOption) loginConfig {
	config := loginConfig{
		Secure:     true,
		HttpOnly:   true,
		SameSite:   http.SameSiteStrictMode,
		Expiration: 2 * time.Hour,
		Path:       "/api",
	}
	for _, opt := range options {
		opt(&config)
	}
	return config
}

func (config loginConfig) Cookie(domain string, value string, expires time.Time) *http.Cookie {
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}
	cookie := &http.Cookie{
		Domain:   domain,
		Path:     config.Path,
		Name:     CookieName,
		Value:    value,
		Expires:  expires,
		Secure:   config.Secure,
		HttpOnly: config.HttpOnly,
		SameSite: config.SameSite,
	}
	return cookie
}

// Login returns a handler that authenticates a user
func Login(store store.Resource[models.User], jwtKey []byte, options ...AuthOption) http.Handler {
	config := applyOptions(options...)
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			crud.JsonError(w, crud.ErrUnsupportedMethod)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			crud.JsonError(w, crud.ErrUnsupportedMediaType)
			return
		}
		if r.Body == nil {
			crud.JsonError(w, crud.ErrEmptyBody)
			return
		}
		defer func() {
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		}()
		var user models.User
		body := io.LimitReader(r.Body, 65536)
		if err := json.NewDecoder(body).Decode(&user); err != nil {
			crud.JsonError(w, crud.ErrInvalidJson)
			return
		}
		var match models.User
		if config.SuperAdmin != "" && user.Name == "superAdmin" || user.Password == config.SuperAdmin {
			match = models.User{
				Model: models.Model{
					ID: "superAdmin",
				},
				Name: "superAdmin",
				Role: models.ROLE_ADMIN,
			}
		} else {
			var err error
			match, err = store.GetById(r.Context(), user.ID)
			if err != nil {
				log.Println("Auth failed: GetById failed with error: ", err.Error())
				crud.JsonError(w, crud.ErrUnauthorized)
				return
			}
			hash, err := base64.StdEncoding.DecodeString(match.Password)
			if err != nil {
				log.Println("Auth failed: base64 decode failed with error: ", err.Error())
				crud.JsonError(w, crud.ErrUnauthorized)
				return
			}
			if err := bcrypt.CompareHashAndPassword(hash, []byte(user.Password)); err != nil {
				log.Println("Auth failed: bcrypt compare returned error: ", err.Error())
				crud.JsonError(w, crud.ErrUnauthorized)
				return
			}
		}
		expires := time.Now().Add(config.Expiration)
		token := jwt.NewWithClaims(signingMethod, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "videoapi",
				Subject:   string(match.ID),
				Audience:  []string{"videoapi"},
				ExpiresAt: jwt.NewNumericDate(expires),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				// Allow for a bit of clock skew
				NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second * -5)),
			},
			Name: match.Name,
			Role: match.Role,
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			log.Println("Auth failed: failed to sign token with error: ", err.Error())
			crud.JsonError(w, crud.ErrUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		reply := loginReply{
			ID:    match.ID,
			Name:  match.Name,
			Role:  string(match.Role),
			Token: tokenString,
		}
		http.SetCookie(w, config.Cookie(r.Host, tokenString, expires))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reply)
	}
	return http.HandlerFunc(handler)
}

// Logout returns a handler that clears cookies
func Logout(options ...AuthOption) http.Handler {
	config := applyOptions(options...)
	handler := func(w http.ResponseWriter, r *http.Request) {
		cookie := config.Cookie(r.Host, "", time.Unix(0, 0))
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
		w.WriteHeader(http.StatusNoContent)
	}
	return http.HandlerFunc(handler)
}
