package entity

import "time"

type Profile struct {
	ID        uint64
	UserID    uint64
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
