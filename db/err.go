package db

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrCanceldDB    = errors.New("processing db canceled")
	ErrDBTimeOut    = errors.New("db timeout")
)
