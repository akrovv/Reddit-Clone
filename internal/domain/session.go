package domain

type Session struct {
	ID   string
	User *User
}

type SessionContextKey string
