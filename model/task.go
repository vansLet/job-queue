package model

import "time"

type Task struct {
	ID       int
	Type     string
	Payload  string
	Status   Status
	CreateAt time.Time
}
