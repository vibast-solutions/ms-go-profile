package entity

import "time"

type Company struct {
	ID             uint64
	Name           string
	RegistrationNo string
	FiscalCode     string
	ProfileID      uint64
	Type           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
