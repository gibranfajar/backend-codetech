package model

import "time"

type About struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AboutRequest struct {
	Title       string `form:"title" validate:"required"`       // required field
	Description string `form:"description" validate:"required"` // required field
}
