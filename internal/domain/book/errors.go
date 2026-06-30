package book

import "errors"

var (
	ErrNotFound        = errors.New("book not found")
	ErrInvalidTitle    = errors.New("book must have a title")
	ErrInvalidStatus   = errors.New("invalid book status")
	ErrPageOutOfRange  = errors.New("page is out of range for this book")
	ErrNoProgress      = errors.New("reading must advance the current page")
	ErrTotalPagesUnset = errors.New("book has no total pages defined")
)
