package domain

import "time"

type Post struct {
	Score            int        `json:"score"`
	Views            uint       `json:"views"`
	Type             string     `json:"type"`
	Title            string     `json:"title"`
	URL              string     `json:"url,omitempty"`
	Author           *Profile   `json:"author"`
	Category         string     `json:"category"`
	Text             string     `json:"text,omitempty"`
	Votes            []*Vote    `json:"votes"`
	Comments         []*Comment `json:"comments"`
	Created          time.Time  `json:"created"`
	UpvotePercentage uint       `json:"upvotePercentage"`
	ID               string     `json:"id"`
}

type Vote struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}
