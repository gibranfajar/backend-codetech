package model

import "time"

type CategoryArticle struct {
	Id        int       `json:"id"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryArticleRequest struct {
	Category string `form:"category" validate:"required"`
}
