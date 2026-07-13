package cache

import "errors"

var ErrInvalidTTL = errors.New("cache TTL must be greater than zero")
