package main

import (
	"encoding/json"
	"os"
	"sync"
)

const fileName = "data.json"

type ShortenReq struct {
	OriginalURL string `json:"url"`
}

type ShortenResp struct {
	ShortURL string `json:"short_url"`
}

type URLData struct {
	OriginalURL string `json:"url"`
	Clicks      int    `json:"click"`
}

type URLStore struct {
	urls map[string]*URLData
	mu   sync.RWMutex
}

var store = &URLStore{
	urls: make(map[string]*URLData),
}

func (s *URLStore) Load() error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) { // First time no file
			return nil
		}
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(file, &s.urls)
}

func (s *URLStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.urls, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, 0o644) // 0644 permission file
}
