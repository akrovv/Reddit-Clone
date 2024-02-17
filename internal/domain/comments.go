package domain

import "time"

type Comment struct {
	Created time.Time `json:"created"`
	Author  *Profile  `json:"author"`
	Body    string    `json:"body"`
	ID      string    `json:"id"`
}
