package models

import "time"

type FMSPC struct {
	FMSPC     string    `json:"fmspc" db:"fmspc"`
	Platform  string    `json:"platform" db:"platform"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type FMSPCResponse struct {
	FMSPC    string `json:"fmspc"`
	Platform string `json:"platform"`
}
