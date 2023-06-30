package main

import (
	"log"
	"net/http"
	"time"
)

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.Printf("HTTP %s %s %s %s", r.RemoteAddr, r.Method, r.URL, time.Since(start))
	})
}
