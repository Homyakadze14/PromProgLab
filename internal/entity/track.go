package entity

import "time"

type Track struct {
	Title    string        `json:"music"`
	Duration time.Duration `json:"duration"`
	Genre    string        `json:"genre"`
	Rating   float32       `json:"rating"`
}
