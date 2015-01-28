package main

import (
	goetcd "github.com/coreos/go-etcd/etcd"
	godocker "github.com/fsouza/go-dockerclient"
	"time"
)

type Deployer struct {
	docker      *godocker.Client
	registry    string
	auth        godocker.AuthConfiguration
	etcd        *goetcd.Client
	etcdPrefix  string
	repoUpdate  chan string
	killTimeout uint
}

func NewDeployer(docker *godocker.Client, registry string, auth godocker.AuthConfiguration, etcd *goetcd.Client, etcdPrefix string, killTimeout uint, pullPeriod time.Duration) *Deployer {
	deployer := &Deployer{
		docker:      docker,
		registry:    registry,
		auth:        auth,
		etcd:        etcd,
		etcdPrefix:  etcdPrefix,
		repoUpdate:  make(chan string, 100),
		killTimeout: killTimeout,
	}

	go deployer.repoUpdateWorker()

	if pullPeriod.Nanoseconds() > 0 {
		go deployer.repoTimerWorker(pullPeriod)
	}

	return deployer
}
