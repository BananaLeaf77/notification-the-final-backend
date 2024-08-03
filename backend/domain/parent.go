package domain

import "time"

type Parent struct {
	ID              int        `json:"id"`
	Name            string     `json:"name" valid:"required~Name is required"`
	Gender          string     `json:"gender" valid:"required~Gender is required"`
	TelephoneNumber int        `json:"telephone_number" valid:"required~Telephone Number is required"`
	Email           string     `json:"email" valid:"email~Invalid email format"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
