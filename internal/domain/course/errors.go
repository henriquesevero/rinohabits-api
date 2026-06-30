package course

import "errors"

var (
	ErrNotFound        = errors.New("course not found")
	ErrInvalidTitle    = errors.New("course must have a title")
	ErrInvalidStatus   = errors.New("invalid course status")
	ErrNoProgress      = errors.New("study session must log positive hours")
	ErrTotalHoursUnset = errors.New("course has no total hours defined")
)
