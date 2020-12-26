package e

import "errors"

var (
	//ErrResourceNotFound resource not exist
	ErrResourceNotFound = errors.New("resource not found")

	//ErrInvalidPortFormat ...
	ErrInvalidPortFormat = errors.New("invalid port format")
)
