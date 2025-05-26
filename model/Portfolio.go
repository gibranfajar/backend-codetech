package model

import "time"

type Portfolio struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PortfolioRequest struct {
	Title string `form:"title" validate:"required"`
	Url   string `form:"url" validate:"required"`
}
