package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"io/ioutil"
)

var (
	ErrCfgNotFound = fmt.Errorf("Dockercfg is missing entry for registry.")
)

type DockerCfg map[string]struct {
	docker.AuthConfiguration
	Auth string `json:auth`
}

func AuthFromDockerCfg(file, registry string) (docker.AuthConfiguration, error) {
	auth := docker.AuthConfiguration{}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return auth, err
	}

	cfg := DockerCfg{}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return auth, err
	}

        r, ok := cfg[registry]
        if !ok {
		return auth, ErrCfgNotFound
	}

	r.ServerAddress = registry
	if r.Auth != "" {
		creds, err := base64.StdEncoding.DecodeString(r.Auth)
		if err != nil {
			return auth, err
		}
		r.Username, r.Password = splitTwo(string(creds), ":")
	}
	return r.AuthConfiguration, nil
}
