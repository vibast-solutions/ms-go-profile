package entity

import "time"

type Address struct {
	ID             uint64
	StreetName     string
	StreenNo       string
	City           string
	County         string
	Country        string
	ProfileID      uint64
	PostalCode     string
	Building       string
	Apartment      string
	AdditionalData string
	Type           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
