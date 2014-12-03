package main

import (
	docker "github.com/fsouza/go-dockerclient"
	"regexp"
)

var (
	IsID = regexp.MustCompile("^[0-9a-f]{12,}$")
)

func (deployer *Deployer) FindStaleContainers() ([]Container, error) {
	apicontainers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

	stale := make([]Container, 0, 5)
	for _, apicontainer := range apicontainers {
		// If image is an ID, it means the tag got reassigned.
		if IsID.MatchString(apicontainer.Image) {
			stale = append(stale, deployer.NewContainer(apicontainer))
		}
	}

	return stale, nil
}

type Container struct {
	deployer *Deployer
	docker.APIContainers
}

func (deployer *Deployer) NewContainer(apicontainer docker.APIContainers) Container {
	container := Container{
		deployer:      deployer,
		APIContainers: apicontainer,
	}
	return container
}

func (container *Container) Restart() error {
	return container.deployer.client.RestartContainer(container.ID, container.deployer.killTimeout)
}

func (container *Container) Stop() error {
	return container.deployer.client.StopContainer(container.ID, container.deployer.killTimeout)
}
