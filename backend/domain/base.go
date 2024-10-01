package domain

import "time"

type SafeStaffData struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type SafeStaffUpdatePayload struct{
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Role      string     `json:"role"`
	UpdatedAt time.Time  `json:"updated_at"`
}
