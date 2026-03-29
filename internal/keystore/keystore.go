package keystore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Key struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	keys map[string]Key // keyed by KeyID (hex)
}

func New(path string) (*Store, error) {
	s := &Store{
		path: path,
		keys: make(map[string]Key),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var keys []Key
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	for _, k := range keys {
		s.keys[k.KeyID] = k
	}
	return nil
}

func (s *Store) save() error {
	keys := make([]Key, 0, len(s.keys))
	for _, k := range s.keys {
		keys = append(keys, k)
	}
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Generate() (Key, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyID := make([]byte, 16)
	key := make([]byte, 16)
	if _, err := rand.Read(keyID); err != nil {
		return Key{}, fmt.Errorf("generate key_id: %w", err)
	}
	if _, err := rand.Read(key); err != nil {
		return Key{}, fmt.Errorf("generate key: %w", err)
	}

	k := Key{
		KeyID: hex.EncodeToString(keyID),
		Key:   hex.EncodeToString(key),
	}
	s.keys[k.KeyID] = k

	if err := s.save(); err != nil {
		return Key{}, err
	}
	return k, nil
}

func (s *Store) Get(keyIDHex string) (Key, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	k, ok := s.keys[keyIDHex]
	return k, ok
}

func (s *Store) List() []Key {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]Key, 0, len(s.keys))
	for _, k := range s.keys {
		keys = append(keys, k)
	}
	return keys
}
