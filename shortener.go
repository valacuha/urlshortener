package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sync"
)

var (
	ErrInvalidURL = errors.New("некорректный URL")
	ErrNotFound   = errors.New("короткий URL не найден")
)

type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// Shorten создает короткий идентификатор для URL
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", fmt.Errorf("%w: %s", ErrInvalidURL, originalURL)
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	var shortID string
	for {
		shortID = generateShortID()
		if _, exists := us.urls[shortID]; !exists {
			break
		}
	}

	us.urls[shortID] = originalURL
	return shortID, nil
}

// GetOriginal возвращает оригинальный URL по короткому ID
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	if shortID == "" {
		return "", ErrNotFound
	}

	us.mu.RLock()
	defer us.mu.RUnlock()

	originalURL, exists := us.urls[shortID]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrNotFound, shortID)
	}

	return originalURL, nil
}

// generateShortID генерирует случайный короткий идентификатор (6 символов)
func generateShortID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:6]
}

// isValidURL проверяет корректность URL
func isValidURL(str string) bool {
	if str == "" {
		return false
	}
	u, err := url.ParseRequestURI(str)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

