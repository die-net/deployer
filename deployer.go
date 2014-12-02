package main

import (
	"flag"
	docker "github.com/fsouza/go-dockerclient"
	"time"
)

var (
	refresh      = flag.Duration("refresh", time.Minute, "Polling frequency of local docker status")
	repull      = flag.Duration("repull", time.Hour * 24, "Polling frequency of remote docker repositories")
)

type Deployer struct {
	client *docker.Client
        registry string
	auth docker.AuthConfiguration

	dockerEvents chan *docker.APIEvents
}

func NewDeployer(client *docker.Client, registry string, auth docker.AuthConfiguration) *Deployer {
	deployer := &Deployer{
		client: client,
		registry: registry,
                auth: auth,
	}

	return deployer
}
