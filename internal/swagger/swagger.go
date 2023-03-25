//go:generate cue export -f -o dist/swagger.yaml swagger.cue
package swagger

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dist/*
var swaggerFS embed.FS
var handler = http.FileServer(http.FS(swaggerFS))

func init() {
	subfs, err := fs.Sub(swaggerFS, "dist")
	if err != nil {
		log.Fatal("Failed to create subfs: ", err)
	}
	handler = http.FileServer(http.FS(subfs))
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
