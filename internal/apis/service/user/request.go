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
