package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleShorten(t *testing.T) {
	shortener := NewURLShortener()
	srv := NewServer(shortener)
	handler := srv.SetupRoutes()

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectShortURL bool
	}{
		{
			name:           "успешное сокращение URL",
			method:         http.MethodPost,
			body:           `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectShortURL: true,
		},
		{
			name:           "некорректный JSON",
			method:         http.MethodPost,
			body:           `{"url": "https://example.com"`,
			expectedStatus: http.StatusBadRequest,
			expectShortURL: false,
		},
		{
			name:           "невалидный URL в JSON",
			method:         http.MethodPost,
			body:           `{"url": "invalid-url"}`,
			expectedStatus: http.StatusBadRequest,
			expectShortURL: false,
		},
		{
			name:           "пустой URL в JSON",
			method:         http.MethodPost,
			body:           `{"url": ""}`,
			expectedStatus: http.StatusBadRequest,
			expectShortURL: false,
		},
		{
			name:           "неверный HTTP метод (GET вместо POST)",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectShortURL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("получен статус %d, ожидался %d. Тело: %s", rec.Code, tt.expectedStatus, rec.Body.String())
			}

			if tt.expectShortURL {
				var resp ShortenResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("не удалось распарсить JSON ответа: %v", err)
				}
				if len(resp.ShortURL) < 6 || len(resp.ShortURL) > 8 {
					t.Errorf("некорректный short_url: %s", resp.ShortURL)
				}
				if resp.OriginalURL != "https://example.com" {
					t.Errorf("некорректный original_url: %s", resp.OriginalURL)
				}
			}
		})
	}
}

func TestHandleRedirect(t *testing.T) {
	shortener := NewURLShortener()
	targetURL := "https://github.com"
	shortID, err := shortener.Shorten(targetURL)
	if err != nil {
		t.Fatalf("ошибка подготовки данных: %v", err)
	}

	srv := NewServer(shortener)
	handler := srv.SetupRoutes()

	tests := []struct {
		name             string
		method           string
		path             string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "успешный редирект по существующему ID",
			method:           http.MethodGet,
			path:             "/" + shortID,
			expectedStatus:   http.StatusFound,
			expectedLocation: targetURL,
		},
		{
			name:             "несуществующий ID возвращает 404",
			method:           http.MethodGet,
			path:             "/noexist",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
		{
			name:             "корневой путь без ID возвращает 404",
			method:           http.MethodGet,
			path:             "/",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
		{
			name:             "неверный HTTP метод (DELETE)",
			method:           http.MethodDelete,
			path:             "/" + shortID,
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("получен статус %d, ожидался %d", rec.Code, tt.expectedStatus)
			}

			if tt.expectedLocation != "" {
				loc := rec.Header().Get("Location")
				if loc != tt.expectedLocation {
					t.Errorf("заголовок Location = %q, ожидался %q", loc, tt.expectedLocation)
				}
			}
		})
	}
}

