package handler

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/chivincent/widevine/internal/keystore"
)

// ClearKey license request/response per W3C EME spec
// https://www.w3.org/TR/encrypted-media/#clear-key-request-format

type clearKeyRequest struct {
	Kids []string `json:"kids"`
	Type string   `json:"type"`
}

type jsonWebKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	K   string `json:"k"`
}

type clearKeyResponse struct {
	Keys []jsonWebKey `json:"keys"`
	Type string       `json:"type"`
}

type LicenseHandler struct {
	store *keystore.Store
}

func NewLicenseHandler(store *keystore.Store) *LicenseHandler {
	return &LicenseHandler{store: store}
}

func (h *LicenseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// if cookie, err := r.Cookie("user"); err != nil || cookie.Value != "fake-user" {
	// 	http.Error(w, "missing cookie", http.StatusUnauthorized)
	// 	return
	// }

	var req clearKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp := clearKeyResponse{Type: "temporary"}

	for _, kidB64 := range req.Kids {
		// kid is base64url-encoded (no padding) 16-byte key ID
		kidBytes, err := base64.RawURLEncoding.DecodeString(kidB64)
		if err != nil {
			log.Printf("license: invalid kid encoding: %s", kidB64)
			continue
		}

		kidHex := hex.EncodeToString(kidBytes)
		k, ok := h.store.Get(kidHex)
		if !ok {
			log.Printf("license: key not found for kid=%s", kidHex)
			continue
		}

		keyBytes, err := hex.DecodeString(k.Key)
		if err != nil {
			log.Printf("license: invalid key hex for kid=%s", kidHex)
			continue
		}

		resp.Keys = append(resp.Keys, jsonWebKey{
			Kty: "oct",
			Kid: kidB64,
			K:   base64.RawURLEncoding.EncodeToString(keyBytes),
		})
	}

	if len(resp.Keys) == 0 {
		http.Error(w, "no keys found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
