package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2"
	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/cors"
	"github.com/warpcomdev/videoapi/internal/crud"
	"github.com/warpcomdev/videoapi/internal/models"
	"github.com/warpcomdev/videoapi/internal/store"
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
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	if len(jwtKey) == 0 {
		log.Fatal("JWT_KEY is not set")
	}
	superAdmin := os.Getenv("SUPER_ADMIN")

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

	db.SetMaxOpenConns(10)                  // his is a small scale server, 10 conns are enough
	db.SetMaxIdleConns(10)                  // defaultMaxIdleConns = 2
	db.SetConnMaxLifetime(30 * time.Minute) // 0, connections are reused forever.

	userDescriptor := models.UserDescriptor()
	if err := userDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", userDescriptor.TableName)
	}
	userResource := store.Adapt[models.User](
		userDescriptor.TableName,
		userDescriptor.FilterSet,
		SqlxQuerier{DB: db},
		SqlxExecutor{DB: db},
		oracleLimiter,
	)

	videoDescriptor := models.VideoDescriptor()
	if err := videoDescriptor.CreateDb(context.Background(), db); err == nil {
		log.Printf("created table %s\n", videoDescriptor.TableName)
	}
	videoResource := store.Adapt[models.Video](
		videoDescriptor.TableName,
		videoDescriptor.FilterSet,
		SqlxQuerier{DB: db},
		SqlxExecutor{DB: db},
		oracleLimiter,
	)

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
	mux.Handle("/api/auth", cors.Allow(auth.Login(userResource.Resource, userResource.Querier, jwtKey, superAdmin)))
	mux.Handle("/api/me", cors.Allow(auth.WithClaims(jwtKey, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ClaimsFrom(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(claims)
	}))))

	// User administration endpoints
	userHandler := http.StripPrefix("/api/user", cors.Allow(auth.WithClaims(jwtKey, crud.Handler(UserRbacResource{
		Resource: userResource,
	}))))
	mux.Handle("/api/user/", userHandler)
	mux.Handle("/api/user", userHandler)

	// video management endpoints
	videoHandler := http.StripPrefix("/api/video", cors.Allow(auth.WithClaims(jwtKey, crud.Handler(RbacResource{
		Resource: videoResource,
	}))))
	mux.Handle("/api/video/", videoHandler)
	mux.Handle("/api/video", videoHandler)

	log.Printf("Listening at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
