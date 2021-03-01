package errors

import (
	"errors"
)

// DBProvider
var ErrInsufficientReadOrWritePermissions = errors.New("insufficient read or write permissions")
var ErrTablesValidationFailed = errors.New("tables validation failed")
var ErrNoActiveEvents = errors.New("no active events")
var ErrMoreThanOneUser = errors.New("more than one user")
var ErrNoUsersFound = errors.New("no users found")
var ErrAlreadySubscribed = errors.New("already subscribed")
var ErrNotSubscribed = errors.New("not subscribed")

// TelegramProvider
var ErrArgumentNotProvided = errors.New("argument not provided")
var ErrInvalidArgument = errors.New("invalid argument")
