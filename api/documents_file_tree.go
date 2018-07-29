package api

import (
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

func DocumentsFileTree(http *transport.HttpClientCtx) *filetree.FileTreeCtx {
	documents := make([]model.Document, 0)

	if err := http.Get(transport.UserBearer, listDocs, nil, &documents); err != nil {
		log.Error.Println("failed to fetch documents %s", err.Error())
		return nil
	}

	fileTree := filetree.CreateFileTreeCtx()

	for _, d := range documents {
		fileTree.AddDocument(d)
	}

	for _, d := range fileTree.Root().Children {
		log.Trace.Println(d.Name(), d.IsFile())
	}

	return &fileTree
}
