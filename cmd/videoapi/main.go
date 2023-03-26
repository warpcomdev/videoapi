package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/rand"

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

	// Readi directory folders
	tmpFolder := os.Getenv("TMPDIR")
	if tmpFolder == "" {
		tmpFolder = "/tmp"
	}
	finalFolder := os.Getenv("FINALDIR")
	if finalFolder == "" {
		panic("FINALDIR must be set")
	}

	// JWT_KEY can be specified for debugging purposes,
	// but it is recommended tolet it generate a random one.
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = make([]byte, 32)
		rand.Read(jwtKey)
	}

	// Couple of debugging aids:
	// - superAdmin password
	// - DEBUG flag to disable security in cookies
	superPassword := os.Getenv("SUPER_PASSWORD")
	debug := strings.HasPrefix(strings.ToLower(os.Getenv("DEBUG")), "t")

	// Read connection strings
	connStr := os.ExpandEnv(os.Args[1])

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
	} else {
		log.Printf("error creating table %s: %s", userDescriptor.TableName, err)
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

	// Camera
	cameraDescriptor := models.CameraDescriptor()
	if err := cameraDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", cameraDescriptor.TableName)
	} else {
		log.Printf("error creating table %s: %s", cameraDescriptor.TableName, err)
	}
	policedCameraStore := policy.CameraPolicy{
		CameraStore: store.New[models.Camera](
			SqlxQuerier{DB: db},
			SqlxExecutor{DB: db},
			cameraDescriptor.TableName,
			cameraDescriptor.FilterSet,
			oracleLimiter,
		),
	}

	// Videos
	videoDescriptor := models.VideoDescriptor()
	if err := videoDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", videoDescriptor.TableName)
	} else {
		log.Printf("error creating table %s: %s", videoDescriptor.TableName, err)
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
		log.Printf("created table %s\n", pictureDescriptor.TableName)
	} else {
		log.Printf("error creating table %s: %s", pictureDescriptor.TableName, err)
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

	// Alerts
	alertDescriptor := models.AlertDescriptor()
	if err := alertDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", alertDescriptor.TableName)
	} else {
		log.Printf("error creating table %s: %s", alertDescriptor.TableName, err)
	}
	policedAlertStore := policy.AlertPolicy{
		AlertStore: store.New[models.Alert](
			SqlxQuerier{DB: db},
			SqlxExecutor{DB: db},
			alertDescriptor.TableName,
			alertDescriptor.FilterSet,
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
	// Camera administration endpoints
	stackHandlers("/api/camera", crud.FromResource(store.Adapt[models.Camera](policedCameraStore)))
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
	// Alert administration endpoints
	stackHandlers("/api/alert", crud.FromResource(store.Adapt[models.Alert](policedAlertStore)))

	// Add swagger UI server
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.HandlerFunc(swagger.ServeHTTP)))

	log.Printf("Listening at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
