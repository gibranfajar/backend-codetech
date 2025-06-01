package model

import "time"

type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Password  string    `json:"password" validate:"required" min:"8"`
	Profile   string    `json:"profile" validate:"required"`
	Role      string    `json:"role" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRequest struct {
	Name     string `form:"name" validate:"required,min=2"`
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=6"`
	Role     string `form:"role" validate:"required"`
}

type UserRequestUpdate struct {
	Name  string `form:"name" validate:"required,min=2"`
	Email string `form:"email" validate:"required,email"`
	Role  string `form:"role" validate:"required"`
}

type UserResponse struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Profile   string    `json:"profile"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
