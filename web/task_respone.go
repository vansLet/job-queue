package web

import "time"

type ResponeTask struct {
	ID       int       `json:"id"`
	Type     string    `json:"type"`
	Payload  string    `json:"payload"`
	Status   string    `json:"status"`
	CreateAt time.Time `json:"create_at"`
}
