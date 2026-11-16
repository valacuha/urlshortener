package main

import (
	"sync"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"валидный HTTP URL", "http://example.com", false},
		{"валидный HTTPS URL", "https://google.com/search?q=test", false},
		{"валидный URL с путем и портом", "https://localhost:8080/api/v1/users", false},
		{"невалидный URL без схемы", "example.com", true},
		{"невалидный URL неверная схема", "ftp://://example.com", true},
		{"невалидный URL произвольный текст", "not-a-url", true},
		{"пустая строка", "", true},
		{"только схема", "https://", true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(shortID) < 6 || len(shortID) > 8 {
					t.Errorf("некорректная длина короткого ID (%d): %s", len(shortID), shortID)
				}
			}
		})
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	shortener := NewURLShortener()
	validURL := "https://golang.org"
	shortID, err := shortener.Shorten(validURL)
	if err != nil {
		t.Fatalf("не удалось подготовить тестовый URL: %v", err)
	}

	tests := []struct {
		name    string
		shortID string
		wantURL string
		wantErr bool
	}{
		{"существующий ID", shortID, validURL, false},
		{"несуществующий ID", "unknown", "", true},
		{"пустой ID", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := shortener.GetOriginal(tt.shortID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("получили URL = %q, ожидали = %q", gotURL, tt.wantURL)
			}
		})
	}
}

func TestURLShortener_UniquenessAndConcurrency(t *testing.T) {
	shortener := NewURLShortener()
	const count = 100

	var wg sync.WaitGroup
	ids := make(chan string, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := shortener.Shorten("https://example.com")
			if err != nil {
				t.Errorf("неожиданная ошибка при конкурентном вызове: %v", err)
				return
			}
			ids <- id
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("обнаружен дубликат короткого ID: %s", id)
		}
		seen[id] = true
	}
}

