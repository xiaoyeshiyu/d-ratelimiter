package rate_limiter

import "github.com/pkg/errors"

var (
	LimitedErr = errors.New("request limited")
)
