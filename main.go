package main

import (
	"context"
	"os"

	"github.com/xlab/treeprint"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/go-unixfs/importer/balanced"
	ihelper "github.com/ipfs/go-unixfs/importer/helpers"
)

var log = logging.Logger("main")

func main() {
	ctx := context.Background()
	logging.SetLogLevel("main", "debug")

	filePath := "/Users/sky/Downloads/Docker.dmg"
	file, err := os.Open(filePath)
	defer file.Close()
	if nil != err {
		log.Errorf("open file err: %s", err.Error())
		return
	}

	chunkerString := "size-262144"
	splitter, err := chunker.FromString(file, chunkerString)
	if err != nil {
		log.Errorf("create splitter err: %s", err.Error())
		return
	}

	dagService := NewDAGService()
	params := ihelper.DagBuilderParams{
		Dagserv:    dagService,
		RawLeaves:  false,
		Maxlinks:   ihelper.DefaultLinksPerBlock,
		NoCopy:     false,
		CidBuilder: cid.V0Builder{},
	}

	dagBuilderHelper, err := params.New(splitter)
	if err != nil {
		log.Errorf("create dagBuilderHelper err: %s", err.Error())
		return
	}

	var node ipld.Node
	//node, err = trickle.Layout(dagBuilderHelper)
	node, err = balanced.Layout(dagBuilderHelper)
	if err != nil {
		log.Errorf("Layout err: %s", err.Error())
		return
	}

	for cid, node := range dagService.Nodes {
		size, err := node.Size()
		if nil != err {
			log.Error(err.Error())
			continue
		}
		if 0 != len(node.Links()) {
			log.Infof("%s:%d:%d", cid.String(), size, len(node.Links()))
		}
	}

	tree := treeprint.NewWithRoot(node)
	walkNode(ctx, dagService, tree, node)

	log.Info(tree.String())
}

func walkNode(ctx context.Context, dagService *DAGService, tree treeprint.Tree, node ipld.Node) {
	for _, link := range node.Links() {
		_node, err := link.GetNode(ctx, dagService)
		if nil != err {
			log.Error(err)
			break
		}
		branch := tree.AddBranch(_node)
		walkNode(ctx, dagService, branch, _node)
	}
}
