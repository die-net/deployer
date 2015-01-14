package main

import (
	"flag"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"runtime"
)

var (
	endpoint     = flag.String("docker", "unix:///var/run/docker.sock", "Docker endpoint to connect to.")
	maxThreads   = flag.Int("max_threads", runtime.NumCPU(), "Maximum number of running threads.")
	registry     = flag.String("registry", "", "URL of docker registry.")
	authUsername = flag.String("auth-username", "", "Username for authentication to docker registry.")
	authPassword = flag.String("auth-password", "", "Password for authentication to docker registry.")
	authEmail    = flag.String("auth-email", "", "Email address for authentication to docker registry.")
	authServer   = flag.String("auth-server", "", "Server address for authentication to docker registry.")
	killTimeout  = flag.Int("kill_timeout", 10, "Container stop timeout, before hard kill (in seconds).")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*maxThreads)

	client, err := docker.NewClient(*endpoint)
	if err != nil {
		log.Fatalln("Couldn't docker.NewClient: ", err)
	}

	auth := docker.AuthConfiguration{
		Username:      *authUsername,
		Password:      *authPassword,
		Email:         *authEmail,
		ServerAddress: *authServer,
	}

	deployer := NewDeployer(client, *registry, auth, uint(*killTimeout))

	deployer.ImageUpdateRepo("")
	deployer.StopStaleContainers()
}
