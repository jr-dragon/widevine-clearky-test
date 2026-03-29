package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chivincent/widevine/internal/handler"
	"github.com/chivincent/widevine/internal/keystore"
	"github.com/chivincent/widevine/internal/packager"
)

func main() {
	videosDir := envOr("VIDEOS_DIR", "./videos")
	keysPath := envOr("KEYS_PATH", "./keys.json")
	addr := envOr("ADDR", ":8080")
	webDir := envOr("WEB_DIR", "./web")

	if err := os.MkdirAll(videosDir, 0755); err != nil {
		log.Fatalf("create videos dir: %v", err)
	}

	store, err := keystore.New(keysPath)
	if err != nil {
		log.Fatalf("init keystore: %v", err)
	}

	pkg := packager.New()

	mux := http.NewServeMux()

	// API routes
	mux.Handle("/api/license", handler.NewLicenseHandler(store))
	mux.Handle("/api/encrypt", handler.NewEncryptHandler(store, pkg, videosDir))
	mux.Handle("/api/videos", handler.NewListHandler(videosDir))

	// Serve encrypted video content with CORS headers for DASH
	videoFS := http.StripPrefix("/videos/", http.FileServer(http.Dir(videosDir)))
	mux.HandleFunc("/videos/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		videoFS.ServeHTTP(w, r)
	})

	// Serve web frontend
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
