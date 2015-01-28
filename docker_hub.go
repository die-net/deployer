package main

import (
	"encoding/json"
	"net/http"
	"path"
	"time"
)

func (deployer *Deployer) RegisterDockerHubWebhook(path string) {
	http.HandleFunc(path, deployer.DockerHubWebhookHandler)

	go deployer.webhookWatchWorker()
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
		deployer.etcd.Set(deployer.etcdPrefix+"/"+repo, time.Now().String(), 0)
	}
}

func (deployer *Deployer) webhookWatchWorker() {
	watch := NewWatch(deployer.etcd, deployer.etcdPrefix, 100)
	for node := range watch.C {
		deployer.repoUpdate <- path.Base(node.Key)
	}
}
