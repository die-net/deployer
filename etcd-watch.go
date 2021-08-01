package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"time"

	goetcd "github.com/coreos/go-etcd/etcd"
)

type Watch struct {
	client     *goetcd.Client
	prefix     string
	watchIndex uint64
	sentIndex  uint64
	C          chan *goetcd.Node
}

func NewWatch(client *goetcd.Client, prefix string, limit int) *Watch {
	watch := &Watch{
		client:     client,
		prefix:     prefix,
		watchIndex: 0,
		sentIndex:  0,
		C:          make(chan *goetcd.Node, limit),
	}

	go watch.worker()

	return watch
}

func (watch *Watch) worker() {
	defer close(watch.C)

	if _, err := watch.client.SetDir(watch.prefix, 0); err != nil {
		// Ignore error code 102 (directory exists).
		var e *goetcd.EtcdError
		if errors.As(err, &e) && e.ErrorCode != 102 {
			log.Println("Watch etcd.SetDir error", watch.prefix, err)
			return
		}
	}

	for {
		// Fetch all current keys under this prefix, recursively.
		resp, err := watch.client.Get(watch.prefix, false, true)
		if err != nil {
			log.Println("Watch etcd.Get error", watch.prefix, err)
			return
		}

		log.Println("Watch etcd.Watch starting index", resp.EtcdIndex, watch.prefix)

		// Start watching for updates after the current watchIndex given in the Get.
		watch.watchIndex = resp.EtcdIndex

		// Find all non-directory nodes and send each to the channel.  This will catch up
		// on any nodes created before we started and any missed during connection retry.
		watch.sendNodes(resp.Node)

		// With strongConsistency, it should be impossible for EtcdIndex
		// to be less than any node.ModifiedIndex or watch.sentIndex.
		if resp.EtcdIndex < watch.sentIndex {
			log.Println("Watch etcd.Watch initial EtcdIndex", resp.EtcdIndex, "less than sentIndex", watch.sentIndex)
		}

		for {
			// Fetch the next changed node for this prefix after watchIndex.
			resp, err = watch.client.Watch(watch.prefix, watch.watchIndex+1, true, nil, nil)
			if err != nil {
				// After 5 minutes, etcd either closes the connection
				// or returns a json.SyntaxError. Retry watch.
				var se *json.SyntaxError
				if errors.Is(err, io.EOF) || errors.As(err, &se) {
					log.Println("Watch etcd.Watch retrying connection", watch.prefix)
					break
				}

				// 401 means our watchIndex is too old, and we need to Get a new one.
				var ee *goetcd.EtcdError
				if errors.As(err, &ee) && ee.ErrorCode == 401 {
					log.Println("Watch etcd.Watch watchIndex", watch.watchIndex+1, "too old", watch.prefix)
					break
				}

				log.Println("Watch etcd.Watch error", watch.prefix, err)
				return
			}

			// Send the changed node(s) to the update channel
			watch.watchIndex = watch.sendNodes(resp.Node)
		}

		time.Sleep(*etcdRetryDelay)
	}
}

func (watch *Watch) sendNodes(node *goetcd.Node) uint64 {
	i := watch.sendNodesRecursively(node)

	// sendNodesRecursively won't encounter nodes in order.  Only update
	// sentIndex when it is done.
	if i > watch.sentIndex {
		watch.sentIndex = i
	}

	return i
}

func (watch *Watch) sendNodesRecursively(node *goetcd.Node) uint64 {
	if !node.Dir {
		// Send this to channel if it is not a repeat.
		if node.ModifiedIndex > watch.sentIndex {
			log.Println("Watch etcd.Watch sendNodes sending ", node)
			node.Key = strings.TrimPrefix(node.Key, watch.prefix+"/")
			watch.C <- node
		}
		return node.ModifiedIndex
	}

	index := node.ModifiedIndex
	for _, node := range node.Nodes {
		// Iterate into directory, track largest index we've seen.
		if i := watch.sendNodesRecursively(node); i > index {
			index = i
		}
	}
	return index
}
