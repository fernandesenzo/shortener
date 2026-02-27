package auth

type loginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}
