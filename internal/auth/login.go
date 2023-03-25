package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

// Login returns a handler that authenticates a user
func Login(store store.Resource[models.User, *models.User], querier store.Querier, jwtKey []byte, superAdmin string) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.Body == nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		defer func() {
			io.Copy(io.Discard, r.Body)
			r.Body.Close()
		}()
		var user models.User
		body := io.LimitReader(r.Body, 65536)
		if err := json.NewDecoder(body).Decode(&user); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var match models.User
		if superAdmin != "" && user.Name == "superAdmin" || user.Password == superAdmin {
			match = models.User{
				Model: models.Model{
					ID: "superAdmin",
				},
				Name: "superAdmin",
				Role: models.ROLE_ADMIN,
			}
		} else {
			var err error
			match, err = store.GetById(r.Context(), querier, user.ID)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			hash, err := base64.StdEncoding.DecodeString(match.Hash)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if err := bcrypt.CompareHashAndPassword(hash, []byte(user.Password)); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		token := jwt.NewWithClaims(signingMethod, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "videoapi",
				Subject:   string(match.ID),
				Audience:  []string{"videoapi"},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
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
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		reply := loginReply{
			ID:    match.ID,
			Name:  match.Name,
			Role:  string(match.Role),
			Token: tokenString,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reply)
	}
	return http.HandlerFunc(handler)
}
