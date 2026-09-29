package model

import "errors"

type Status int

var nameStatus []string = []string{
	"Pending",
	"Processing",
	"Complete",
	"Failed",
}

const (
	Pending Status = iota
	Processing
	Complete
	Failed
)

func ToStatus(s string) (Status, error) {
	for i := range len(nameStatus) {
		if s == nameStatus[i] {
			return Status(i), nil
		}
	}
	return -1, errors.New("invalid status")
}

func (s Status) String() string {
	i := int(s)
	if i < len(nameStatus) {
		return nameStatus[i]
	}
	return nameStatus[3]
}
