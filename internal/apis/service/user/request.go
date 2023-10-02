package user

type CreateAccountRequest struct {
	Provider string `json:"provider"`
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Provider string `json:"provider"`
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GetQueryRequest struct {
	ID string `json:"id"`
}

type UpdateEmailRequest struct {
	NewEmail string `json:"new_email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
