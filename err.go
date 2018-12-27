package main

import (
	"errors"
)

var (
	ErrNoTag = errors.New("docker image tag not specified")
)
