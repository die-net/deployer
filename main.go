package main

import (
	"flag"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"runtime"
)

var (
	endpoint    = flag.String("docker", "unix:///var/run/docker.sock", "Docker endpoint to connect to.")
	maxThreads  = flag.Int("max_threads", runtime.NumCPU(), "Maximum number of running threads.")
	registry    = flag.String("registry", "https://index.docker.io/v1/", "URL of docker registry.")
	dockerCfg   = flag.String("dockercfg", os.Getenv("HOME")+"/.dockercfg", "Path to .dockercfg authentication information.")
	killTimeout = flag.Int("kill_timeout", 10, "Container stop timeout, before hard kill (in seconds).")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*maxThreads)

	client, err := docker.NewClient(*endpoint)
	if err != nil {
		log.Fatalln("Couldn't docker.NewClient: ", err)
	}

	auth := docker.AuthConfiguration{}
	if *dockerCfg != "" {
		auth, err = AuthFromDockerCfg(*dockerCfg, *registry)
		if err != nil {
			log.Fatalln("AuthFromDockerCfg: ", err)
		}
	}

	deployer := NewDeployer(client, *registry, auth, uint(*killTimeout))

	deployer.ImageUpdateRepo("")

	deployer.StopStaleContainers()
}
