package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	ErrCfgNotFound = fmt.Errorf("dockercfg is missing entry for registry")
)

// DockerCfg is the undocumented format for ~/.dockercfg file.
type DockerCfg map[string]struct {
	docker.AuthConfiguration        // Decode most fields directly into our return format.
	Auth                     string `json:"auth"` // Allow "auth" field too, so we can convert.
}

func AuthFromDockerCfg(file, registry string) (docker.AuthConfiguration, error) {
	empty := docker.AuthConfiguration{}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return empty, err
	}

	cfgs := DockerCfg{}
	if err := json.Unmarshal(content, &cfgs); err != nil {
		return empty, err
	}

	cfg, ok := cfgs[registry]
	if !ok {
		return empty, ErrCfgNotFound
	}

	// Split apart base64 "auth" field into Username and Password.
	if cfg.Auth != "" {
		creds, err := base64.StdEncoding.DecodeString(cfg.Auth)
		if err != nil {
			return empty, err
		}
		cfg.Username, cfg.Password = splitTwo(string(creds), ":")
	}

	cfg.ServerAddress = registry

	return cfg.AuthConfiguration, nil
}
