Deployer [![Build Status](https://travis-ci.org/die-net/deployer.svg?branch=master)](https://travis-ci.org/die-net/deployer)
========

A simple Go-based continuous deployment system, based on catching webhooks from the Docker Hub build service.

Goals
-----

* Workflow based on continuous deployment.  When a build succeeds, deploy as soon as possible.
* Don't leave different versions of the same app running for any longer than necessary.
* Avoid polling against the docker repository frequently.
* Avoid fleet's race conditions by using docker directly. (Interoperate with fleet, but don't depend on it.)
* Zero-downtime deploys.  Only one instance of a container should be offline at a time. (Not implemented.)

Workflow
--------

* Github is configured to send webhooks to the Docker Hub build service, kicking off a build.
* Deployer runs on each machine in the cluster
* One instance gets a [webhook](https://docs.docker.com/docker-hub/builds/#webhooks) from the Docker Hub build service, indicating that a repo has a new build
* Deployer writes repo and timestamp to etcd.
* The deployer on each machine is watching etcd, and sees that a given repo has been updated.
* It uses [go-dockerclient](https://github.com/fsouza/go-dockerclient) to ask Docker which image repotags we have on the local machine. It starts a docker pull for each image found belonging to a given repotag.
* When the docker pull is complete, it compares the list of running containers against the list of images.  If any running containers have a newer image available, they are sent a ```docker stop``` at which point the fleet unit restarts the app using the new image.
* Slack is optionally notified that the deploy is done.

Using Deployer
--------------

Deployer assumes you are using Docker with Fleet and etcd, and that you've set all of your Fleet units to auto-restart.  For example if your company is called MyOrg and app is called Foo, and you have a foo.service file that looks something like:

```
[Unit]
Description=Foo App

[Service]
User=core
EnvironmentFile=/etc/environment.myorg
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=-/usr/bin/docker pull myorg/foo:${MYORG_ENV}
ExecStart=/usr/bin/docker run --name %n -m 2g -P --env-file=/etc/environment.myorg myorg/foo:${MYORG_ENV}
ExecStop=/usr/bin/docker stop %n
TimeoutStartSec=5m
Restart=always
StartLimitInterval=5m

[X-Fleet]
Global=true
```

This fleet unit assumes that you've written some environment variables to /etc/environment.myorg that vary based on whether you are in production, staging, qa, development, etc, that you have Docker tags corresponding to the names of those environments.  For example, MYORG_ENV could be set to "qa" and thus your full Docker repotag would be: ```myorg/foo:qa```

There's a pre-built [Docker Hub](https://hub.docker.com/r/dienet/deployer/) image available as ```dienet/deployer:latest```, which you could run with a Fleet unit like:


```
[Unit]
Description=Deployer

[Service]
User=core
EnvironmentFile=/etc/environment.myorg
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=-/usr/bin/docker pull dienet/deployer:latest
ExecStart=/usr/bin/docker run --name %n -m 1g -h ${HOSTNAME} -P -v /var/run/docker.sock:/var/run/docker.sock -v /home/core/.dockercfg:/app/.dockercfg dienet/deployer:latest --dockercfg /app/.dockercfg --etcd_nodes=http://${ETCD} --etcd_prefix=/${MYORG_ENV}/good/deployer --slack_username="${MYORG_ENV} deploy" --slack_webhook_url=https://hooks.slack.com/services/...
ExecStop=/usr/bin/docker stop %n
TimeoutStartSec=5m
Restart=always
RestartSec=1m

[X-Fleet]
Global=true
```

The above fleet unit assumes you've also written the .dockercfg credentials file to /home/core/.dockercfg (perhaps by running ```docker login```).  It maps ```/var/run/docker.sock``` into the container so deployer has access to the docker API on the local machine.

Set Github to notify Docker Hub when anything is committed to ```myorg/foo```, Docker Hub kicks off a build, which then is configured to send a webhook to a qa deployer URL when the build completes.

All commandline flags:

```
-docker string
      Docker endpoint to connect to. (default "unix:///var/run/docker.sock")
-dockercfg string
      Path to .dockercfg authentication information. (default "~/.dockercfg")
-etcd_dial_timeout duration
      How long to wait to connect to etcd nodes. (default 5s)
-etcd_nodes string
      Comma-seperated list of etcd nodes to connect to. (default "http://127.0.0.1:4001")
-etcd_prefix string
      Path prefix for etcd nodes. (default "/deployer")
-etcd_retry_delay duration
      How long to between request retries. (default 2s)
-kill_timeout int
      Container stop timeout, before hard kill (in seconds). (default 10)
-listen string
      [IP]:port to listen for incoming connections. (default ":4500")
-max_threads int
      Maximum number of running threads. (default 4)
-registry string
      URL of docker registry. (default "https://index.docker.io/v1/")
-repull_period duration
      How frequently to re-pull all images, without any notification. (default 24h0m0s)
-slack_username string
      Username to show in slack. (default "Deployer")
-slack_webhook_url string
      Slack incoming webhook url (optional)
-strong_consistency
      Set etcd consistency level as strong.
-webhook_path string
      Path to webhook from Docker Hub. (default "/api/dockerhub/webhook")
```

License
-------

Copyright 2014, 2015, 2016 Aaron Hopkins and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
