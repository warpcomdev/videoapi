package main

import (
	"context"
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
	tmpFolder := os.Getenv("TMPDIR")
	if tmpFolder == "" {
		tmpFolder = "/tmp"
	}
	finalFolder := os.Getenv("FINALDIR")
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

	// Create policed stores for every crud resource
	// Users
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

	// Videos
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

	// Pictures
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
	mux.Handle("/api/me", cors.Allow(auth.WithClaims(jwtKey, http.HandlerFunc(handleMe))))

	// Stack all the cors, auth and crud middleware on top of the resources
	stackHandlers := func(prefix string, frontend crud.Frontend) {
		handler := http.StripPrefix(prefix, cors.Allow(auth.WithClaims(jwtKey, crud.NewHandler(frontend))))
		mux.Handle(prefix+"/", handler)
		mux.Handle(prefix, handler)
	}

	// User administration endpoints
	stackHandlers("/api/user", crud.FromResource(store.Adapt[models.User](policedUserStore)))
	// Video administration endpoints
	stackHandlers("/api/video", crud.FromMedia(
		store.Adapt[models.Video](policedVideoStore),
		tmpFolder,
		finalFolder,
		map[string]string{
			"video/4gpp":      ".4gpp",
			"video/3gpp2":     ".3gpp2",
			"video/3gp2":      ".3gp2",
			"video/mpeg":      ".mpg",
			"video/mp4":       ".mp4",
			"video/ogg":       ".ogg",
			"video/quicktime": ".quicktime",
			"video/webm":      ".webm",
		},
	))
	// Picture administration endpoints
	stackHandlers("/api/picture", crud.FromMedia(
		store.Adapt[models.Picture](policedPictureStore),
		tmpFolder,
		finalFolder,
		map[string]string{
			"image/jpeg": ".jpg",
			"image/png":  ".png",
		},
	))

	// Add swagger UI server
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.HandlerFunc(swagger.ServeHTTP)))

	log.Printf("Listening at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
