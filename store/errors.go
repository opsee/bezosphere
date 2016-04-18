package store

import (
	"errors"
)

var (
	errMissingCustomerId      = errors.New("missing customer id")
	errMissingAWSRequest      = errors.New("missing AWS request struct")
	errMissingAWSOutput       = errors.New("missing AWS resource output")
	errMissingAWSRequestInput = errors.New("missing AWS request input")
	errMissingUpdated         = errors.New("missing updated_at timestamp")
	errResourceExpired        = errors.New("cached resource has expired")
)
