package errors

import "errors"

// Errors for persistent storage
var ErrPersistentStorageEmpty = errors.New("persistent storage empty")
var ErrPersistentStorageNotEmpty = errors.New("persistent storage not empty")

// Errors for subscriptions
var ErrSubscriptionNotExist = errors.New("subscription does not exist")
var ErrAlreadySubscribed = errors.New("already subscribed")
var ErrNotSubscribed = errors.New("not subscribed")

// Errors for users
var ErrUserNotFound = errors.New("user not found")
