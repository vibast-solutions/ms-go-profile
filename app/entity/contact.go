package entity

import "time"

type Contact struct {
	ID        uint64
	FirstName string
	LastName  string
	NIN       string
	DOB       *time.Time
	Phone     string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	ProfileID uint64
}
