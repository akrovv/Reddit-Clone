package domain

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ID       string `json:"id"`
}

type Profile struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}
