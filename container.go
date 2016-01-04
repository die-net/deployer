package main

import (
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"regexp"
	"sort"
	"strings"
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

func (deployer *Deployer) InspectContainer(id string) (*docker.Container, error) {
	return deployer.docker.InspectContainer(id)
}

func (deployer *Deployer) StopContainers(containers []docker.APIContainers) {
	names := []string{}

	for _, container := range containers {
		log.Println("Stopping container", container.ID, container.Names)
		names = append(names, container.Names...)

		err := deployer.docker.StopContainer(container.ID, deployer.killTimeout)
		if err != nil {
			log.Println("Stop container", err)
		}
	}

	if slack != nil && len(names) > 0 {
		sort.Strings(names)
		// Container Names are prefixed by "/".
		text := "Deploying " + strings.Replace(strings.TrimPrefix(strings.Join(names, " "), "/"), " /", ", ", -1)
		if err := slack.Send(SlackPayload{Text: text}); err != nil {
			log.Println("Slack error: ", err)
		}
	}
}

func (deployer *Deployer) FindStaleContainers() ([]docker.APIContainers, error) {
	repotagMap, err := deployer.ListRepotags()
	if err != nil {
		return nil, err
	}

	filter := func(apicontainer *docker.APIContainers) bool {
		// If image is an ID, it means the tag got reassigned.
		if IsID.MatchString(apicontainer.Image) {
			return true
		}

		// Otherwise, apicontainer.ID is a repotag.  Make sure image container is running is still current.
		if container, err := deployer.InspectContainer(apicontainer.ID); err == nil && repotagMap[apicontainer.Image] != container.Image {
			return true
		}

		return false
	}
	return deployer.FindContainers(docker.ListContainersOptions{}, filter)
}

func (deployer *Deployer) StopStaleContainers() {
	containers, err := deployer.FindStaleContainers()
	if err != nil {
		log.Println("FindStaleContainers", err)
		return
	}

	deployer.StopContainers(containers)
}
