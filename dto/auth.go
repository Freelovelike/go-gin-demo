package dto

// RegisterRequest 是用户注册的请求主体。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// LoginRequest 是用户登录的请求主体。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 在成功登录/注册后返回。
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse 是公开的用户信息（没有密码）。
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Gold     int    `json:"gold"`
	Level    int    `json:"level"`
	ExpVal   int    `json:"exp_val"`
	ExpToLvl int    `json:"exp_to_lvl"`
}
