package service

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user with this name or email already exists")
	ErrInvalidInput          = errors.New("invalid input provided")
	ErrJWTGenerationFailed   = errors.New("failed to generate token")
	ErrTokenNotFound         = errors.New(" data not set or empty")
	ErrTokenExpiredOrInvalid = errors.New("token has expired or invalid")
	ErrPasswordMismatch      = errors.New("passwords do not match")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrInternalServerError   = errors.New("internal server error")
	ErrTooManyRequests       = errors.New("too many requests, please try again later")
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrTokenRevoked          = errors.New("token has been revoked")
)
