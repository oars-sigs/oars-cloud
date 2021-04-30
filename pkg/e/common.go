package e

import "errors"

var (
	//ErrResourceNotFound resource not exist
	ErrResourceNotFound = errors.New("resource not found")

	//ErrInvalidPortFormat ...
	ErrInvalidPortFormat = errors.New("invalid port format")

	//ErrResourceExisted ...
	ErrResourceExisted = errors.New("resource had existed")

	//ErrCACertNotFound ...
	ErrCACertNotFound = errors.New("ca cert not found")

	//ErrInvalidContainerDrive ...
	ErrInvalidContainerDrive = errors.New("invalid container drive")
)
