package d_ratelimiter

import "github.com/pkg/errors"

var (
	LimitedErr = errors.New("request limited")
)
