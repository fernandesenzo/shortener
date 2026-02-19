package domain

import "errors"

// link errors
var ErrLinkNotFound = errors.New("link not found")
var ErrInvalidURL = errors.New("invalid URL")
var ErrURLTooLong = errors.New("URL too long")
var ErrLinkCreationFailed = errors.New("link creation failed")
var ErrUserExceededLinkLimit = errors.New("user already has too many links saved")

// user errors
var ErrNicknameAlreadyUsed = errors.New("nickname already exists")
var ErrPasswordTooLong = errors.New("password too long")
var ErrPasswordTooShort = errors.New("password too short")
