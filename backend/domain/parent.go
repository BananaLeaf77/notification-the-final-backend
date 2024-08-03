package domain

import "time"

type Parent struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Gender          string     `json:"gender"`
	TelephoneNumber int        `json:"telephone_number"`
	Email           string     `json:"email"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
