package main

import (
	"encoding/json"
	goetcd "github.com/coreos/go-etcd/etcd"
	"io"
	"log"
	"strings"
	"time"
)

type Watch struct {
	client    *goetcd.Client
	prefix    string
	sentIndex uint64
	C         chan *goetcd.Node
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
	defer close(watch.C)

	if _, err := watch.client.SetDir(watch.prefix, 0); err != nil {
		// Ignore error code 102 (directory exists).
		if e, ok := err.(*goetcd.EtcdError); !ok || e.ErrorCode != 102 {
			log.Println("Watch etcd.SetDir error", watch.prefix, err)
			return
		}
	}

	for {
		// Fetch all current keys under this prefix, recursively.
		resp, err := watch.client.Get(watch.prefix, true, true)
		if err != nil {
			log.Println("Watch etcd.Get error", watch.prefix, err)
			return
		}

		// Start watching for updates after the current index given in the Get.
		index := resp.EtcdIndex

		log.Println("Watch etcd.Watch starting index", index, watch.prefix)

		// Find all non-directory nodes and send each to the channel.  This will catch up
		// on any nodes created before we started and any missed during connection retry.
		if i := watch.sendNodes(resp.Node); i > watch.sentIndex {
			watch.sentIndex = i
		}

		for {
			// Fetch the next changed node for this prefix after index.
			resp, err = watch.client.Watch(watch.prefix, index+1, true, nil, nil)
			if err != nil {
				// After 5 minutes, etcd either closes the connection
				// or returns a json.SyntaxError. Retry watch.
				if _, ok := err.(*json.SyntaxError); ok || err == io.EOF {
					log.Println("Watch etcd.Watch retrying connection", watch.prefix)
					break
				}

				// 401 means our index is too old, and we need to Get a new one.
				if e, ok := err.(*goetcd.EtcdError); ok && e.ErrorCode == 401 {
					log.Println("Watch etcd.Watch index", index+1, "too old", watch.prefix)
					break
				}

				log.Println("Watch etcd.Watch error", watch.prefix, err)
				return
			}

			// Send the changed node(s) to the update channel, track largest index we've sent.
			if i := watch.sendNodes(resp.Node); i > index {
				if i < watch.sentIndex {
					log.Println("Watch etcd.Watch bug: found old nodes again")
					return
				}
				index = i
				watch.sentIndex = i
			}
		}

		time.Sleep(*etcdRetryDelay)
	}
}

func (watch *Watch) sendNodes(node *goetcd.Node) uint64 {
	if !node.Dir {
		// Send this to channel if it is not a repeat.
		if node.ModifiedIndex > watch.sentIndex {
			log.Println("Watch etcd.Watch sendNodes sending ", node)
			node.Key = strings.TrimPrefix(node.Key, watch.prefix+"/")
			watch.C <- node
		}
		return node.ModifiedIndex
	}

	var index uint64
	for _, node := range node.Nodes {
		// Iterate into directory, track largest index we've seen.
		if i := watch.sendNodes(node); i > index {
			index = i
		}
	}
	return index
}
