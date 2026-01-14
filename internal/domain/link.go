package domain

import "time"

type Link struct {
	ID          string
	OriginalURL string
	ShortURL    string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}
