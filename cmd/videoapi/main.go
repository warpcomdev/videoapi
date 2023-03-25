package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2"
	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/cors"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/policy"
	"github.com/warpcomdev/videoapi/internal/store"
	"github.com/warpcomdev/videoapi/internal/swagger"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

// response to /me enpoint
type meResponse struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Role    models.Role `json:"role"`
	Expires time.Time   `json:"expires"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("\nhello_ora")
		fmt.Println("\thello_ora check if it can connect to the given oracle server, then print server banner.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("\thello_ora oracle://user:pass@server/service_name")
		fmt.Println()
		os.Exit(1)
	}
	connStr := os.ExpandEnv(os.Args[1])
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	if len(jwtKey) == 0 {
		log.Fatal("JWT_KEY is not set")
	}
	superPassword := os.Getenv("SUPER_PASSWORD")
	debug := strings.HasPrefix(strings.ToLower(os.Getenv("DEBUG")), "t")

	db, err := sqlx.Connect("oracle", connStr)
	for {
		if err == nil {
			break
		}
		log.Println("Can't create connection:", err)
		log.Println("Sleeping for ten seconds")
		time.Sleep(10 * time.Second)
		db, err = sqlx.Connect("oracle", connStr)
	}

	db.SetMaxOpenConns(10)                  // this is a small scale server, 10 conns are enough
	db.SetMaxIdleConns(10)                  // defaultMaxIdleConns = 2
	db.SetConnMaxLifetime(30 * time.Minute) // if 0, connections are reused forever.

	userDescriptor := models.UserDescriptor()
	if err := userDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", userDescriptor.TableName)
	}
	// Need access to the unpoliced UserStore for login
	userStore := store.New[models.User](
		SqlxQuerier{DB: db},
		SqlxExecutor{DB: db},
		userDescriptor.TableName,
		userDescriptor.FilterSet,
		oracleLimiter,
	)
	policedUserStore := policy.UserPolicy{
		UserStore: userStore,
	}

	videoDescriptor := models.VideoDescriptor()
	if err := videoDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", videoDescriptor.TableName)
	}
	policedVideoStore := policy.VideoPolicy{
		VideoStore: store.New[models.Video](
			SqlxQuerier{DB: db},
			SqlxExecutor{DB: db},
			videoDescriptor.TableName,
			videoDescriptor.FilterSet,
			oracleLimiter,
		),
	}

	pictureDescriptor := models.PictureDescriptor()
	if err := pictureDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", videoDescriptor.TableName)
	}
	policedPictureStore := policy.PicturePolicy{
		PictureStore: store.New[models.Picture](
			SqlxQuerier{DB: db},
			SqlxExecutor{DB: db},
			pictureDescriptor.TableName,
			pictureDescriptor.FilterSet,
			oracleLimiter,
		),
	}

	mux := &http.ServeMux{}
	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    65535,
	}

	// Authorization endpoints
	authOptions := make([]auth.AuthOption, 0, 8)
	if superPassword != "" {
		authOptions = append(authOptions, auth.WithSuperAdmin(superPassword))
	}
	if debug {
		authOptions = append(authOptions,
			auth.WithSecureCookie(false),
			auth.WithSameSiteCookie(false),
		)
	}
	mux.Handle("/api/login", cors.Allow(auth.Login(userStore, jwtKey, authOptions...)))
	mux.Handle("/api/logout", cors.Allow(auth.Logout(authOptions...)))
	mux.Handle("/api/me", cors.Allow(auth.WithClaims(jwtKey, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ClaimsFrom(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := meResponse{
			ID:      claims.Subject,
			Name:    claims.Name,
			Role:    claims.Role,
			Expires: claims.ExpiresAt.Time,
		}
		json.NewEncoder(w).Encode(resp)
	}))))

	// Stack all the cors, auth and crud middleware on top of the resources
	stackHandlers := func(prefix string, resource crud.Resource) {
		handler := http.StripPrefix(prefix, cors.Allow(auth.WithClaims(jwtKey, crud.Handler(resource))))
		mux.Handle(prefix+"/", handler)
		mux.Handle(prefix, handler)
	}

	// User administration endpoints
	stackHandlers("/api/user", store.Adapt[models.User](policedUserStore))
	stackHandlers("/api/picture", store.Adapt[models.Picture](policedPictureStore))
	stackHandlers("/api/video", store.Adapt[models.Video](policedVideoStore))

	// Add swagger UI server
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.HandlerFunc(swagger.ServeHTTP)))

	log.Printf("Listening at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
