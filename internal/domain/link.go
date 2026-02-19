package domain

import "time"

type Link interface {
	GetCode() string
	GetOriginalURL() string
}
type PermanentLink struct {
	ID          string
	OriginalURL string
	Code        string
	UserID      string
	CreatedAt   time.Time
}

func (p PermanentLink) GetCode() string        { return p.Code }
func (p PermanentLink) GetOriginalURL() string { return p.OriginalURL }

type TemporaryLink struct {
	OriginalURL string
	Code        string
}

func (t TemporaryLink) GetCode() string        { return t.Code }
func (t TemporaryLink) GetOriginalURL() string { return t.OriginalURL }
