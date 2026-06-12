package dto

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginRequest is the request body for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful login/register.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is the public user info (no password).
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Gold     int    `json:"gold"`
	Level    int    `json:"level"`
	ExpVal   int    `json:"exp_val"`
	ExpToLvl int    `json:"exp_to_lvl"`
}
