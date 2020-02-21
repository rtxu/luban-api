package db

import (
	"errors"
)

var (
	// errors could be handled
	ErrNotFound = errors.New(`no more rows in this result set`)
)
