package main

import (
	"errors"
)

var (
	ErrNoTag          = errors.New("Docker image tag not specified")
	ErrAlreadyPulling = errors.New("Docker repotag already being pulled")
)
