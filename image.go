package main

// Handle docker pull requests.  Don't allow duplicate outstanding pull
// requests for the same repo tag, and track start time for each pull
// request, for future logging/display.

import (
	"log"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

// FindRepoTags finds valid repotags in the list of local images.  Accepts 3 forms:
// "*"        - Match any repotag.
// "foo"      - Match any repotag in the repo "foo".
// "foo:bar"  - Exact match the repotag "foo:bar" if it exists.
func (deployer *Deployer) FindRepoTags(repo string) ([]string, error) {
	images, err := deployer.docker.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		return nil, err
	}

	prefix := false
	if repo != "*" && !strings.Contains(repo, ":") {
		repo += ":"
		prefix = true
	}
	repotags := make([]string, 0, 5)

	for _, image := range images {
		for _, repotag := range image.RepoTags {
			if repotag == "<none>:<none>" {
				continue
			}
			if repo == "*" || repo == repotag {
				repotags = append(repotags, repotag)
			} else if prefix && strings.HasPrefix(repotag, repo) {
				repotags = append(repotags, repotag)
			}
		}
	}

	return repotags, nil
}

func (deployer *Deployer) ListRepotags() (map[string]string, error) {
	images, err := deployer.docker.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		return nil, err
	}

	repotagMap := make(map[string]string)
	for _, image := range images {
		for _, repotag := range image.RepoTags {
			if repotag == "<none>:<none>" {
				continue
			}
			repotagMap[repotag] = image.ID
		}
	}

	return repotagMap, nil
}

func (deployer *Deployer) repoTimerWorker(period time.Duration) {
	tick := time.NewTicker(period)
	for {
		deployer.repoUpdate <- "*" // Update all.
		<-tick.C
	}
}

func (deployer *Deployer) repoUpdateWorker() {
	for repo := range deployer.repoUpdate {
		_ = deployer.ImageUpdateRepo(repo)
		deployer.StopStaleContainers()
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

	log.Println("Pulling image", repotag)

	err := deployer.docker.PullImage(opts, deployer.auth)
	if err != nil {
		log.Println("Pull image error", repotag, err)
	}
	return err
}
