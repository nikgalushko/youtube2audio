package config

import (
	"sync/atomic"
)

const (
	converter = "ffmpeg"
	api       = "http_api"
)

type Node struct {
	Adress string
	Name   string
}

type Nodes struct {
	index *int64
	size  int64
	nodes []Node
}

func (n *Nodes) Add(node Node) {
	if n.index == nil {
		var zero int64
		n.index = &zero
	}
	n.nodes = append(n.nodes, node)
	n.size++
}

// Next return next node from list
func (n *Nodes) Next() Node {
	i := atomic.AddInt64(n.index, 1) % n.size
	nextNode := n.nodes[i]

	return nextNode
}
