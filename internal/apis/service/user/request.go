package user

// RequestCreateAccount is request of POST /user/auth/createAccount
type RequestCreateAccount struct {
	Provider string `json:"provider"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RequestLogin is request of POST /user/auth/login
type RequestLogin struct {
	Provider string `json:"provider"`
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RequestUpdateUserEmail is request of PUT /user/email
type RequestUpdateEmail struct {
	NewEmail string `json:"new_email"`
}

// RequestUpdateUserPassword is request of PUT /user/password
type RequestUpdatePassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
