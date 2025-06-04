package entity

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional: "user" (default) or "admin"
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type EmailVerification struct {
	SessionID string `json:"session_id"`
	Email     string `json:"email"`
	Code      string `json:"code"`
	ExpiresAt int64  `json:"expires_at"`
}

type VerifyEmailDTO struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}
