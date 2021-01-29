package errors

import (
	"errors"
)

// DBProvider
var ErrInsufficientReadOrWritePermissions = errors.New("insufficient read or write permissions")
var ErrTablesValidationFailed = errors.New("tables validation failed")
var ErrNoActiveEvents = errors.New("no active events")
