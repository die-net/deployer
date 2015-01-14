package main

// Handle docker pull requests.  Don't allow duplicate outstanding pull
// requests for the same repo tag, and track start time for each pull
// request, for future logging/display.

import (
	docker "github.com/fsouza/go-dockerclient"
	"strings"
)

func (deployer *Deployer) FindRepoTags(repo string) ([]string, error) {
	images, err := deployer.client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		return nil, err
	}

	if repo != "" {
		repo = repo + ":"
	}
	repotags := make([]string, 0, 5)

	for _, image := range images {
		for _, repotag := range image.RepoTags {
			if repo == "" || strings.HasPrefix(repotag, repo) {
				repotags = append(repotags, repotag)
			}
		}
	}

	return repotags, nil
}

func (deployer *Deployer) repoUpdateWorker() {
	for repo := range deployer.repoUpdate {
		deployer.ImageUpdateRepo(repo)
	}
}

func (deployer *Deployer) ImageUpdateRepo(repo string) error {
	repotags, err := deployer.FindRepoTags(repo)
	if err != nil {
		return err
	}

	return deployer.PullImages(repotags)
}

func (deployer *Deployer) PullImages(repotags []string) error {
	var ret error
	for _, repotag := range repotags {
		if err := deployer.PullImage(repotag); err != nil {
			ret = err
		}
	}
	return ret
}

func (deployer *Deployer) PullImage(repotag string) error {
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
