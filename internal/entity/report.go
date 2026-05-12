package entity

import "time"

type Report struct {
	CountTracks    int           `json:"count_tracks"`
	CommonDuration time.Duration `json:"common_duration"`
	AvgRating      float32       `json:"average rating"`
}
