package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Server struct {
	shortener *URLShortener
}

func NewServer(shortener *URLShortener) *Server {
	return &Server{shortener: shortener}
}

func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", s.handleShorten)
	mux.HandleFunc("/", s.handleRedirect)
	return mux
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "метод не поддерживается"})
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "некорректный JSON"})
		return
	}

	shortID, err := s.shortener.Shorten(req.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp := ShortenResponse{
		ShortURL:    shortID,
		OriginalURL: req.URL,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "метод не поддерживается"})
		return
	}

	shortID := strings.TrimPrefix(r.URL.Path, "/")
	if shortID == "" || strings.Contains(shortID, "/") {
		http.NotFound(w, r)
		return
	}

	originalURL, err := s.shortener.GetOriginal(shortID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func main() {
	shortener := NewURLShortener()
	srv := NewServer(shortener)

	log.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", srv.SetupRoutes()); err != nil {
		log.Fatal(err)
	}
}

