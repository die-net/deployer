package main

import (
	goetcd "github.com/coreos/go-etcd/etcd"
)

type Watch struct {
	client *goetcd.Client
	prefix string
	C      chan *goetcd.Node
}

func NewWatch(client *goetcd.Client, prefix string, limit int) *Watch {
	watch := &Watch{
		client: client,
		prefix: prefix,
		C:      make(chan *goetcd.Node, limit),
	}

	go watch.worker()

	return watch
}

func (watch *Watch) worker() {
	// Fetch all current keys under this prefix, recursively.
	resp, err := watch.client.Get(watch.prefix, true, true)
	if err != nil {
		close(watch.C)
		return
	}

	// Find all non-directory nodes and send each to the channel.
	watch.sendNodes(resp.Node)

	// Start watching for updates after the initial update.
	index := resp.EtcdIndex

	for {
		// Fetch the next changed node for this prefix after index.
		resp, err = watch.client.Watch(watch.prefix, index+1, true, nil, nil)
		if err != nil {
			close(watch.C)
			return
		}

		// Send the changed node(s) to the update channel.
		watch.sendNodes(resp.Node)

		// Watch again for changes after this.
		index = resp.Node.ModifiedIndex
	}
}

func (watch *Watch) sendNodes(node *goetcd.Node) {
	if !node.Dir {
		watch.C <- node
		return
	}

	for _, node := range node.Nodes {
		watch.sendNodes(node)
	}
}
