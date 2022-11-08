package myErrors

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
var ErrClientNotExists = errors.New("client not exists")
var ErrNoTeamBounded = errors.New("no team bounded")
var ErrMoreThenOneTeamBounded = errors.New("more then one team bounded")
var ErrUserAlreadyExists = errors.New("user already exists")

// TelegramProvider
var ErrArgumentNotProvided = errors.New("argument not provided")
var ErrInvalidArgument = errors.New("invalid argument")

// Config
var ErrOTRSLoginNotProvided = errors.New("otrs login not provided")
var ErrOTRSPasswordNotProvided = errors.New("otrs password not provided")
var ErrTelegramTokenNotProvided = errors.New("telegram token not provided")
var ErrMandatoryFieldMissing = errors.New("mandatory fields missing")
