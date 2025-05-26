package model

import "time"

type Faq struct {
	Id         int       `json:"id"`
	Question   string    `json:"question"`
	Answer     string    `json:"answer"`
	CategoryId string    `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FaqRequest struct {
	Question   string `form:"question" validate:"required"`
	Answer     string `form:"answer" validate:"required"`
	CategoryId string `form:"category_id" validate:"required"`
}

type FaqResponse struct {
	Id        int       `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
