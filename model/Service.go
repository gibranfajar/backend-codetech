package model

import "time"

type Service struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceRequest struct {
	Title       string `form:"title" validate:"required"`       // required field
	Description string `form:"description" validate:"required"` // required field
}
