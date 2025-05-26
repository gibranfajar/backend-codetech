package model

import "time"

type Contact struct {
	Id              int       `json:"id"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	Address         string    `json:"address"`
	OfficeOperation string    `json:"office_operation"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ContactRequest struct {
	Phone           string `form:"phone" validate:"required"`            // required field
	Email           string `form:"email" validate:"required"`            // required field
	Address         string `form:"address" validate:"required"`          // required field
	OfficeOperation string `form:"office_operation" validate:"required"` // required field
}
