package model

import "time"

type Article struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	UserId      int       `json:"user_id"`
	CategoryId  int       `json:"category_id"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ResponseArticle struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	User        string    `json:"user"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	Views       int       `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ArticleRequest struct {
	Title       string `form:"title" validate:"required"`
	Description string `form:"description" validate:"required"`
	CategoryId  int    `form:"category_id" validate:"required"` // Add CategoryId field for article creation
	UserId      int    `form:"user_id" validate:"required"`     // Add UserId field for article creation
}
