package main

import (
	"encoding/json"
	"net/http"
)

type DockerHubWebhook struct {
	CallbackURL string              `json:"callback_url"`
	Repository  DockerHubRepository `json:"repository"`
}

type DockerHubRepository struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Owner     string `json:"owner"`
	RepoName  string `json:"repo_name"`
}

func DockerHubWebhookHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var webhook DockerHubWebhook
	if err := decoder.Decode(&webhook); err != nil {
	}

	//	repo := webhook.Repository.RepoName
}

func init() {
	http.HandleFunc("/api/dockerhub/", DockerHubWebhookHandler)
}
