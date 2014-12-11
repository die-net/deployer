package main

import (
	docker "github.com/fsouza/go-dockerclient"
	"regexp"
)

var (
	IsID = regexp.MustCompile("^[0-9a-f]{12,}$")
)

// A filter func returns true if the container should be added to list.
type FilterContainer func(apicontainer *docker.APIContainers) bool

func (deployer *Deployer) FindContainers(options docker.ListContainersOptions, filter FilterContainer) ([]Container, error) {
	apicontainers, err := deployer.client.ListContainers(options)
	if err != nil {
		return nil, err
	}

	stale := make([]Container, 0, 5)
	for _, apicontainer := range apicontainers {
		if filter(&apicontainer) {
			stale = append(stale, deployer.NewContainer(apicontainer))
		}
	}

	return stale, nil
}

func (deployer *Deployer) FindStaleContainers() ([]Container, error) {
	filter := func(apicontainer *docker.APIContainers) bool {
		// If image is an ID, it means the tag got reassigned.
		return IsID.MatchString(apicontainer.Image)
	}
	return deployer.FindContainers(docker.ListContainersOptions{}, filter)
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
