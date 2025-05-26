package model

import "time"

type Product struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	Discount    int       `json:"discount"`
	Type        string    `json:"type"`
	Icon        string    `json:"icon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductRequest struct {
	Title       string `form:"title" validate:"required"`
	Description string `form:"description" validate:"required"`
	Price       int    `form:"price" validate:"required"`
	Type        string `form:"type" validate:"required"`
}
