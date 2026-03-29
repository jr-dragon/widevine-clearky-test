package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type ListHandler struct {
	videosDir string
}

func NewListHandler(videosDir string) *ListHandler {
	return &ListHandler{videosDir: videosDir}
}

type videoEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Manifest string `json:"manifest"`
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(h.videosDir)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var videos []videoEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mpdPath := filepath.Join(h.videosDir, e.Name(), "manifest.mpd")
		if _, err := os.Stat(mpdPath); err != nil {
			continue
		}

		name := e.Name()
		metaPath := filepath.Join(h.videosDir, e.Name(), "meta.json")
		if data, err := os.ReadFile(metaPath); err == nil {
			var meta map[string]string
			if json.Unmarshal(data, &meta) == nil {
				if n, ok := meta["name"]; ok {
					name = n
				}
			}
		}

		videos = append(videos, videoEntry{
			ID:       e.Name(),
			Name:     name,
			Manifest: "/videos/" + e.Name() + "/manifest.mpd",
		})
	}

	if videos == nil {
		videos = []videoEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}
