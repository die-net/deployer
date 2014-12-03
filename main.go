package main

import (
	"flag"
	"fmt"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"runtime"
)

var (
	endpoint   = flag.String("docker", "unix:///var/run/docker.sock", "Docker endpoint to connect to.")
	maxThreads = flag.Int("max_threads", runtime.NumCPU(), "Maximum number of running threads.")
	client     *docker.Client
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*maxThreads)

	var err error
	client, err = docker.NewClient(*endpoint)
	if err != nil {
		log.Fatalln("Couldn't docker.NewClient: ", err)
	}

	deployer := NewDeployer(client, "", docker.AuthConfiguration{}, 10)
	stale, err := deployer.FindStaleContainers()
	if err != nil {
		log.Fatalln(err)
	}
	for _, container := range stale {
		fmt.Printf("Stopping container %s\n", container.ID)
		container.Stop()
		if err != nil {
			fmt.Errorf("Stop container error: %+v\n", err)
		}
	}
}
