package main

import (
	"errors"
)

var (
	ErrNoTag = errors.New("Docker image tag not specified")
)
