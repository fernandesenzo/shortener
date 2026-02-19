package user

type createRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type createResponse struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
}
