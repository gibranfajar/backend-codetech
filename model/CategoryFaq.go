package model

import "time"

type CategoryFaq struct {
	Id          int       `json:"id"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryFaqRequest struct {
	Category    string `form:"category" validate:"required"`
	Description string `form:"description" validate:"required"`
}
