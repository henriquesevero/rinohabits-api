package habit

import "errors"

var (
	ErrNotFound   = errors.New("habit not found")
	ErrNoSchedule = errors.New("habit must have a schedule (weekdays or weekly frequency)")
)
