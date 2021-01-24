package errors

import (
	"errors"
)

// DBProvider
var ErrInsufficientReadOrWritePermissions = errors.New("insufficient read or write permissions")
