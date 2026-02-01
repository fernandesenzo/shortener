package domain

import "time"

type Link struct {
	ID          string
	OriginalURL string
	Code        string
	CreatedAt   time.Time
}
