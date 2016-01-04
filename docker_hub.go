package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
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
		RepoURL  string `json:"repo_url"`
	} `json:"repository"`
}

type DockerHubCallback struct {
	State       string `json:"state"`
	Description string `json:"description"`
	Context     string `json:"context"`
	TargetURL   string `json:"target_url"`
}

func (deployer *Deployer) DockerHubWebhookHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var webhook DockerHubWebhook
	if err := decoder.Decode(&webhook); err != nil {
		return
	}

	repo := webhook.Repository.RepoName
	repoURL := webhook.Repository.RepoURL
	log.Println("Webhook received for", repo, repoURL)

	now := time.Now().String()

	callback := DockerHubCallback{
		State:       "error",
		Description: "An unknown error occurred.",
		Context:     "etcd=" + *etcdNodes + " time=" + now,
		TargetURL:   "http://" + req.Host + "/",
	}

	if repo == "" {
		callback.Description = "Error: webhook JSON repository.repo_name is empty."
	} else {
		key := deployer.etcdPrefix + "/" + repo
		if _, err := deployer.etcd.Set(key, now, 0); err == nil {
			callback.State = "success"
			callback.Description = "Triggered etcd " + key
		} else {
			log.Println("Webhook couldn't etcd.Set", key, err)
			callback.Description = "Error: etcd couldn't set " + key + ": " + err.Error()
		}
	}

	deployer.DockerHubDoCallback(webhook.CallbackURL, callback)
}

func (deployer *Deployer) DockerHubDoCallback(callbackURL string, callback DockerHubCallback) {
	if callbackURL == "" {
		log.Println("webhook.CallbackURL not specified")
		return
	}

	jsonCallback, err := json.Marshal(callback)
	if err != nil {
		log.Println("Webhook couldn't json.Marshal", err)
		return
	}

	resp, err := http.Post(callbackURL, "application/json", bytes.NewBuffer(jsonCallback))
	if err != nil {
		log.Println("Webhook callback POST error:", err)
		return
	}

	// Discard body
	resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Webhook callback POST status:", resp.StatusCode)
	}
}

func (deployer *Deployer) webhookWatchWorker() {
	watch := NewWatch(deployer.etcd, deployer.etcdPrefix, 100)
	for node := range watch.C {
		log.Println("Etcd watch received for", node.Key)
		deployer.repoUpdate <- node.Key
	}
	log.Fatalln("Etcd watcher died.")
}
