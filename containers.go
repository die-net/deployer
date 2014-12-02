package main

import (
	docker "github.com/fsouza/go-dockerclient"
	"regexp"
)

var (
        IsStaleImage = regexp.MustCompile("^[0-9a-f]{12,}$")
)


func (deployer *Deployer) FindStaleContainers() ([]string, error) {
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

        stale := make([]string, 0, 5)
        for _, container := range containers {
                if IsStaleImage.MatchString(container.Image) {
                        stale = append(stale, container.ID)
                }
        }

        return stale, nil
}
