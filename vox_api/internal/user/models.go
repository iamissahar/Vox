package user

type UserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
}

type u struct {
	ID string `json:"user_id"`
}
