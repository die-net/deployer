package main

import (
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"regexp"
)

var (
	IsID = regexp.MustCompile("^[0-9a-f]{12,}$")
)

// A filter func returns true if the container should be added to list.
type FilterContainer func(apicontainer *docker.APIContainers) bool

func (deployer *Deployer) FindContainers(options docker.ListContainersOptions, filter FilterContainer) ([]docker.APIContainers, error) {
	apicontainers, err := deployer.docker.ListContainers(options)
	if err != nil {
		return nil, err
	}

	ret := make([]docker.APIContainers, 0, 5)
	for _, apicontainer := range apicontainers {
		if filter(&apicontainer) {
			ret = append(ret, apicontainer)
		}
	}

	return ret, nil
}

func (deployer *Deployer) StopContainers(containers []docker.APIContainers) {
	for _, container := range containers {
		log.Println("Stopping container", container.ID, container.Names)
		err := deployer.docker.StopContainer(container.ID, deployer.killTimeout)
		if err != nil {
			log.Println("Stop container", err)
		}
	}
}

func (deployer *Deployer) FindStaleContainers() ([]docker.APIContainers, error) {
	filter := func(apicontainer *docker.APIContainers) bool {
		// If image is an ID, it means the tag got reassigned.
		return IsID.MatchString(apicontainer.Image)
	}
	return deployer.FindContainers(docker.ListContainersOptions{}, filter)
}

func (deployer *Deployer) StopStaleContainers() {
	containers, err := deployer.FindStaleContainers()
	if err != nil {
		log.Println("FindStaleContainers", err)
	} else {
		deployer.StopContainers(containers)
	}
}
