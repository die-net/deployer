package main

import (
	"log"
	"regexp"
	"sort"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

var IsID = regexp.MustCompile("^[0-9a-f]{12,}$")

// FilterContainer is a filter func returns true if the container should be added to list.
type FilterContainer func(apicontainer *docker.APIContainers) bool

func (deployer *Deployer) FindContainers(options docker.ListContainersOptions, filter FilterContainer) ([]docker.APIContainers, error) {
	apicontainers, err := deployer.docker.ListContainers(options)
	if err != nil {
		return nil, err
	}

	ret := make([]docker.APIContainers, 0, len(apicontainers))
	for i := range apicontainers {
		apicontainer := &apicontainers[i]
		if filter(apicontainer) {
			ret = append(ret, *apicontainer)
		}
	}

	return ret, nil
}

func (deployer *Deployer) InspectContainer(id string) (*docker.Container, error) {
	return deployer.docker.InspectContainerWithOptions(docker.InspectContainerOptions{ID: id})
}

func (deployer *Deployer) StopContainers(containers []docker.APIContainers) {
	names := []string{}

	for i := range containers {
		container := &containers[i]

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
		text := "Deploying " + strings.ReplaceAll(strings.TrimPrefix(strings.Join(names, " "), "/"), " /", ", ")
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
