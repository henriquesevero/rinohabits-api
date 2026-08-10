package course

import "errors"

var (
	ErrNotFound            = errors.New("course not found")
	ErrInvalidTitle        = errors.New("course must have a title")
	ErrInvalidStatus       = errors.New("invalid course status")
	ErrNoProgress          = errors.New("study session must log positive hours")
	ErrTotalHoursUnset     = errors.New("course has no total hours defined")
	ErrInvalidWeekday      = errors.New("active weekdays must be between 1 and 7")
	ErrInvalidReminderTime = errors.New("reminder time must include a valid hour and minute")
)
