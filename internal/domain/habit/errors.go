package habit

import "errors"

var (
	ErrNotFound        = errors.New("habit not found")
	ErrNoActiveWeekday = errors.New("habit must have at least one active weekday")
)
