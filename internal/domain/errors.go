package domain

import "errors"

var ErrLinkNotFound = errors.New("link not found")
var ErrInvalidURL = errors.New("invalid URL")
var ErrURLTooLong = errors.New("URL too long")
var ErrLinkCreationFailed = errors.New("link creation failed")
