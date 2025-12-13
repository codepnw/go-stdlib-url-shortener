package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const port = 8080

func main() {
	// Load data.json
	if err := store.Load(); err != nil {
		fmt.Printf("Loading data failed: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", shortenHandler)
	mux.HandleFunc("GET /{id}", redirectHandler)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Server running on %s", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Server error: %v", err)
		}
	}()

	// Wait for Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\n Shutting down server..")

	// Graceful Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Shutdown error: ", err)
	}

	// Save data
	if err := store.Save(); err != nil {
		fmt.Println("Error save data: ", err)
	} else {
		fmt.Println("Data saved successfully")
	}

	fmt.Println("Bye!")
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allow", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Generate Short ID
	shortID := generateShortID(6)

	newData := &URLData{
		OriginalURL: req.OriginalURL,
		Clicks:      0,
	}

	// Add URL
	store.mu.Lock()
	store.urls[shortID] = newData
	store.mu.Unlock()

	// Save URL
	store.Save()

	resp := &ShortenResp{
		ShortURL: fmt.Sprintf("http://localhost:%d/%s", port, shortID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var targetURL string
	var exists bool

	func() {
		store.mu.Lock()
		defer store.mu.Unlock()

		if data, ok := store.urls[id]; ok {
			targetURL = data.OriginalURL
			exists = true
			data.Clicks++
		}
	}()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, targetURL, http.StatusFound)
}
