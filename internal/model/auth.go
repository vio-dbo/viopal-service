package model

// Register
type RegisterRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required"`
	PhoneNumber  string `json:"phone_number" validate:"required"`
	Password     string `json:"password" validate:"required"`
	BusinessName string `json:"business_name" validate:"required"`
}

type RegisterResponse struct {
	Register *RegisterRequest `json:"data,omitempty"`
	Error    *ErrorBody       `json:"error,omitempty"`
}

// Login
type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Login *LoginRequest `json:"data,omitempty"`
	Error *ErrorBody    `json:"error,omitempty"`
	Token string        `json:"token"`
}

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	RoleID       int
	Role         string
	BusinessName string
	PhoneNumber  string
}

// Me or Profile
type Me struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	BusinessName string `json:"business_name"`
	PhoneNumber  string `json:"phone_number"`
}

type MeResponse struct {
	Me    *Me        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

type Merchant struct {
	ID           int     `json:"id"`
	UserID       string  `json:"user_id"`
	BusinessName string  `json:"bussiness_name"`
	PhoneNumber  string  `json:"phone_number"`
	Balance      float64 `json:"balance"`
}
