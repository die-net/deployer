package main

// Handle docker pull requests.  Don't allow duplicate outstanding pull
// requests for the same repo tag, and track start time for each pull
// request, for future logging/display.

import (
	docker "github.com/fsouza/go-dockerclient"
	"strings"
)

func (deployer *Deployer) FindRepoTags(repo string) ([]string, error) {
        images, err := client.ListImages(false)
        if err != nil {  
                return nil, err
        }

        repo = repo + ":"
        repotags := make([]string, 0, 5)

        for _, image := range images {
                for _, repotag := range image.RepoTags {
			if strings.HasPrefix(repotag, repo) {
				repotags = append(repotags, repotag)
			}
		}
	}

	return repotags, nil
}


func (deployer *Deployer) ImageUpdateRepo(repo string) error {
        repotags, err := deployer.FindRepoTags(repo)
        if err != nil {
		return err
	}

        for _, repotag := range repotags {
		deployer.ImagePull(repotag)
	}
	return nil
}

func (deployer *Deployer) ImagePull(repotag string) error {
	repo, tag := splitTwo(repotag, ":")
	if tag == "" {
		return ErrNoTag
	}

	opts := docker.PullImageOptions{
		Repository: repo,
		Registry:   deployer.registry,
		Tag:        tag,
	}

	return deployer.client.PullImage(opts, deployer.auth)
}
