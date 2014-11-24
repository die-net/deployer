# Deployer

A sketch of an automated deployment system, based on catching webhooks from the Docker Hub build service.

## Goals

* Workflow based on continuous deployment.  When a build succeeds, deploy as soon as possible.
* Zero-downtime deploys.  Only one instance of a container should be offline at a time.
* Don't leave different versions of the same app running for any longer than necessary.
* Avoid polling against the docker repository frequently.
* Avoid fleet's race conditions by using docker directly. (Interoperate with fleet, but don't depend on it.)

## Non-Goals

* Don't try to solve the local development workflow yet.

## Detailed Design

* Written in Go
* Running on each machine in the cluster
* Using [go-dockerclient](https://github.com/fsouza/go-dockerclient) to communicate with Docker
* Getting [webhooks](https://docs.docker.com/docker-hub/builds/#webhooks) from the Docker Hub build service
* Tracking of the most-recent version of each repo+tag in etcd.
* Only allow one restart at a time via a lock in etcd.

I'm thinking of the app as a sequence of channels.

The external triggers are:

* Build callback webhook: Send forcePull=repository[].repo_name to update channel.

* Docker events watcher: On docker "untag" event, send forcePull=nil to update channel.

* etcd watcher: When another instance of deployer writes to list of images maintained in etcd, send forcePull=nil to the update channel.

* Hourly timer: Send forcePull=nil to image list channel.

### Update channel

1. Issue list-images. For a given repo+tag, compare result with etcd:
    *  etcd has lower CreatedAt: update etcd
    *  etcd hash higher CreatedAt or repo=forcePull: send to image-pull channel
3. Issue list-containers:
    *  for each Image =~ /^[0-9a-f]{12,}$/
    *    Send Names[0] to Restart channel

### Image-pull channel

* Given a repo and tag, issue docker pull.  On success, send forcePull=nil to update channel.
  
### Restart channel

* Acquire etcd restart lock
* Issue docker restart for a given container
* Wait for state to remain "Up" and TCP connection to succeed to container's port, if listed.
* Drop lock
