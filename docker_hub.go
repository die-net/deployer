package main

import (
	"encoding/json"
	"net/http"
)

func (deployer *Deployer) RegisterDockerHubWebhook(path string) {
	http.HandleFunc(path, deployer.DockerHubWebhookHandler)
}

type DockerHubWebhook struct {
	Repository struct {
		RepoName string `json:"repo_name"`
	} `json:"repository"`
}

func (deployer *Deployer) DockerHubWebhookHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var webhook DockerHubWebhook
	if err := decoder.Decode(&webhook); err != nil {
	}

	repo := webhook.Repository.RepoName
	if repo != "" {
		deployer.repoUpdate <- repo
	}
}
