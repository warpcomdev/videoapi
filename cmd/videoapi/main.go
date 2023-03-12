package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2"
	"github.com/warpcomdev/videoapi/crud"
	"github.com/warpcomdev/videoapi/models"
	"github.com/warpcomdev/videoapi/store"
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

	db, err := sqlx.Connect("oracle", connStr)
	if err != nil {
		dieOnError("Can't create connection:", err)
	}
	db.SetMaxOpenConns(10)                  // his is a small scale server, 10 conns are enough
	db.SetMaxIdleConns(10)                  // defaultMaxIdleConns = 2
	db.SetConnMaxLifetime(30 * time.Minute) // 0, connections are reused forever.

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

	videoHandler := http.StripPrefix("/api/video", crud.Handler(videoResource))
	mux.Handle("/api/video/", videoHandler)
	mux.Handle("/api/video", videoHandler)

	log.Printf("Listening at %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
