package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"time"
)

func (deployer *Deployer) RegisterDockerHubWebhook(path string) {
	http.HandleFunc(path, deployer.DockerHubWebhookHandler)

	go deployer.webhookWatchWorker()
}

type DockerHubWebhook struct {
	CallbackURL string `json:"callback_url"`
	Repository  struct {
		RepoName string `json:"repo_name"`
	} `json:"repository"`
}

func (deployer *Deployer) DockerHubWebhookHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var webhook DockerHubWebhook
	if err := decoder.Decode(&webhook); err != nil {
		return
	}

	repo := webhook.Repository.RepoName
	log.Println("Webhook received for", repo)

	if repo != "" {
		key := deployer.etcdPrefix + "/" + repo
		if _, err := deployer.etcd.Set(key, time.Now().String(), 0); err != nil {
			log.Println("Webhook couldn't etcd.Set", key, err)
		}
	}
}

func (deployer *Deployer) webhookWatchWorker() {
	watch := NewWatch(deployer.etcd, deployer.etcdPrefix, 100)
	for node := range watch.C {
		repo := path.Base(node.Key)
		log.Println("Etcd watch received for", repo)
		deployer.repoUpdate <- repo
	}
}
