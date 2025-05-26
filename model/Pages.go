package model

import "time"

type Pages struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Banner      string    `json:"banner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PageRequest struct {
	Title       string `form:"title" validate:"required"`
	Type        string `form:"type" validate:"required"`
	Description string `form:"description" validate:"required"`
}
