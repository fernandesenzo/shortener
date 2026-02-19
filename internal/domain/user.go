package domain

import "time"

type User struct {
	ID           string
	Nickname     string
	PasswordHash string
	CreatedAt    time.Time
}
