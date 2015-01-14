package main

import (
	"flag"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"os"
	"runtime"
	"time"
)

var (
	endpoint     = flag.String("docker", "unix:///var/run/docker.sock", "Docker endpoint to connect to.")
	maxThreads   = flag.Int("max_threads", runtime.NumCPU(), "Maximum number of running threads.")
	registry     = flag.String("registry", "https://index.docker.io/v1/", "URL of docker registry.")
	dockerCfg    = flag.String("dockercfg", os.Getenv("HOME")+"/.dockercfg", "Path to .dockercfg authentication information.")
	killTimeout  = flag.Int("kill_timeout", 10, "Container stop timeout, before hard kill (in seconds).")
	repullPeriod = flag.Duration("refresh_period", 24*time.Hour, "How frequently to re-pull all images, without any notification.")
	webhookPath  = flag.String("webhook_path", "/api/dockerhub/webhook", "Path to webhook from Docker Hub.")
	listenAddr   = flag.String("listen", ":4500", "[IP]:port to listen for incoming connections.")
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

	deployer := NewDeployer(client, *registry, auth, uint(*killTimeout), *repullPeriod)
	deployer.RegisterDockerHubWebhook(*webhookPath)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
