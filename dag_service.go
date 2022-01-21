package main

import (
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
)

type DAGService struct {
	mu    sync.Mutex
	Nodes map[cid.Cid]format.Node
}

func NewDAGService() *DAGService {
	return &DAGService{Nodes: make(map[cid.Cid]format.Node)}
}

func (d *DAGService) Get(ctx context.Context, cid cid.Cid) (format.Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n, ok := d.Nodes[cid]; ok {
		return n, nil
	}
	return nil, format.ErrNotFound
}

func (d *DAGService) GetMany(ctx context.Context, cids []cid.Cid) <-chan *format.NodeOption {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make(chan *format.NodeOption, len(cids))
	for _, c := range cids {
		if n, ok := d.Nodes[c]; ok {
			out <- &format.NodeOption{Node: n}
		} else {
			out <- &format.NodeOption{Err: format.ErrNotFound}
		}
	}
	close(out)
	return out
}

func (d *DAGService) Add(ctx context.Context, node format.Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Nodes[node.Cid()] = node
	return nil
}

func (d *DAGService) AddMany(ctx context.Context, nodes []format.Node) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, n := range nodes {
		d.Nodes[n.Cid()] = n
	}
	return nil
}

func (d *DAGService) Remove(ctx context.Context, c cid.Cid) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.Nodes, c)
	return nil
}

func (d *DAGService) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, c := range cids {
		delete(d.Nodes, c)
	}
	return nil
}
