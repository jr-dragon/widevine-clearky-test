package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chivincent/widevine/internal/keystore"
	"github.com/chivincent/widevine/internal/packager"
)

type EncryptHandler struct {
	store     *keystore.Store
	packager  *packager.Packager
	videosDir string
}

func NewEncryptHandler(store *keystore.Store, pkg *packager.Packager, videosDir string) *EncryptHandler {
	return &EncryptHandler{
		store:     store,
		packager:  pkg,
		videosDir: videosDir,
	}
}

type encryptResponse struct {
	ID       string `json:"id"`
	KeyID    string `json:"key_id"`
	Manifest string `json:"manifest"`
}

func (h *EncryptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Max 500MB upload
	r.ParseMultipartForm(500 << 20)

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "missing video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate key pair
	key, err := h.store.Generate()
	if err != nil {
		log.Printf("encrypt: generate key: %v", err)
		http.Error(w, "failed to generate key", http.StatusInternalServerError)
		return
	}

	// Use key_id as directory name
	outputDir := filepath.Join(h.videosDir, key.KeyID)

	// Save uploaded file temporarily
	name := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	tmpFile := filepath.Join(h.videosDir, fmt.Sprintf("tmp_%s_%s.mp4", key.KeyID, name))
	dst, err := os.Create(tmpFile)
	if err != nil {
		log.Printf("encrypt: create temp file: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		os.Remove(tmpFile)
		log.Printf("encrypt: save upload: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dst.Close()
	defer os.Remove(tmpFile)

	// Run Shaka Packager
	err = h.packager.Encrypt(packager.EncryptRequest{
		InputPath: tmpFile,
		OutputDir: outputDir,
		KeyIDHex:  key.KeyID,
		KeyHex:    key.Key,
	})
	if err != nil {
		log.Printf("encrypt: packager: %v", err)
		http.Error(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	// Save metadata
	meta := map[string]string{
		"name":   name,
		"key_id": key.KeyID,
	}
	metaData, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(outputDir, "meta.json"), metaData, 0644)

	resp := encryptResponse{
		ID:       key.KeyID,
		KeyID:    key.KeyID,
		Manifest: fmt.Sprintf("/videos/%s/manifest.mpd", key.KeyID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
