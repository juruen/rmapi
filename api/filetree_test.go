package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DIRECTORY = true
	FILE      = false
)

func createEntry(id, parent, name string, collection bool) Document {
	var t string

	if collection {
		t = "CollectionType"
	} else {
		t = "DocumentType"
	}

	return Document{ID: id, Parent: parent, VissibleName: name, Type: t}
}

func createFile(id, parent, name string) Document {
	return createEntry(id, parent, name, FILE)
}

func createDirectory(id, parent, name string) Document {
	return createEntry(id, parent, name, DIRECTORY)
}

func TestCreateFileTreeCtx(t *testing.T) {
	ctx := CreateFileTreeCtx()

	assert.Equal(t, "/", ctx.root.Name())
}

func TestAddSingleFileToRoot(t *testing.T) {
	ctx := CreateFileTreeCtx()

	file := createEntry("1", "", "file", false)

	ctx.AddDocument(file)

	assert.Equal(t, 1, len(ctx.root.Children))
	assert.Equal(t, "file", ctx.root.Children["1"].Name())
}

func TestAddDirAndFiles(t *testing.T) {
	ctx := CreateFileTreeCtx()

	dir := createDirectory("1", "", "dir")
	file := createFile("2", "1", "file")
	file1 := createFile("3", "1", "file1")

	ctx.AddDocument(file)
	assert.Equal(t, 1, len(ctx.pendingParent))

	ctx.AddDocument(dir)
	assert.Equal(t, 0, len(ctx.pendingParent))

	ctx.AddDocument(file1)
	assert.Equal(t, 0, len(ctx.pendingParent))

	assert.Equal(t, "/", ctx.root.Name())
	assert.Equal(t, "dir", ctx.root.Children["1"].Name())
	assert.Equal(t, "file", ctx.root.Children["1"].Children["2"].Name())
	assert.Equal(t, "file1", ctx.root.Children["1"].Children["3"].Name())

}

func TestAddSeveralFilesAndDirs(t *testing.T) {
	ctx := CreateFileTreeCtx()

	// dir1/dir12/file1
	// dir2/file2
	// dir3/file3
	// dir3/file4
	// file5.pdf

	dir1 := createDirectory("1", "", "dir1")
	dir12 := createDirectory("2", "1", "dir12")
	dir2 := createDirectory("3", "", "dir2")
	dir3 := createDirectory("4", "", "dir3")

	file1 := createFile("5", "2", "file1")
	file2 := createFile("6", "3", "file2")
	file3 := createFile("7", "4", "file3")
	file4 := createFile("8", "4", "file4")
	file5 := createFile("9", "", "file5")

	ctx.AddDocument(file1)
	ctx.AddDocument(file2)
	ctx.AddDocument(file3)
	ctx.AddDocument(file4)
	ctx.AddDocument(file5)
	ctx.AddDocument(dir3)
	ctx.AddDocument(dir2)
	ctx.AddDocument(dir12)
	ctx.AddDocument(dir1)

	assert.Equal(t, "/", ctx.root.Name())
	assert.Equal(t, "dir1", ctx.root.Children["1"].Name())
	assert.Equal(t, "dir12", ctx.root.Children["1"].Children["2"].Name())
	assert.Equal(t, "file1", ctx.root.Children["1"].Children["2"].Children["5"].Name())
	assert.Equal(t, "file2", ctx.root.Children["3"].Children["6"].Name())
	assert.Equal(t, "file3", ctx.root.Children["4"].Children["7"].Name())
	assert.Equal(t, "file4", ctx.root.Children["4"].Children["8"].Name())
	assert.Equal(t, "file5", ctx.root.Children["9"].Name())
}

func TestNodeByPath(t *testing.T) {
	ctx := CreateFileTreeCtx()

	// dir1/dir12/file1
	dir1 := createDirectory("1", "", "dir1")
	dir12 := createDirectory("2", "1", "dir12")
	file1 := createFile("3", "2", "file1")

	ctx.AddDocument(file1)
	ctx.AddDocument(dir12)
	ctx.AddDocument(dir1)

	node, _ := ctx.NodeByPath("/", ctx.Root())
	assert.Equal(t, "/", node.Name())

	node, _ = ctx.NodeByPath("/dir1", ctx.Root())
	assert.Equal(t, "dir1", node.Name())

	node, _ = ctx.NodeByPath("dir1", ctx.Root())
	assert.Equal(t, "dir1", node.Name())

	node, _ = ctx.NodeByPath("/dir1/dir12", ctx.Root())
	assert.Equal(t, "dir12", node.Name())

	node, _ = ctx.NodeByPath("/dir1/dir12/file1", ctx.Root())
	assert.Equal(t, "file1", node.Name())

	dir12Node, _ := ctx.NodeByPath("/dir1/dir12", ctx.Root())
	node, _ = ctx.NodeByPath("file1", dir12Node)
	assert.Equal(t, "file1", node.Name())

	node, _ = ctx.NodeByPath("../dir12/file1", dir12Node)
	assert.Equal(t, "file1", node.Name())

	node, _ = ctx.NodeByPath("./file1", dir12Node)
	assert.Equal(t, "file1", node.Name())
}

func TestNodeToPath(t *testing.T) {
	ctx := CreateFileTreeCtx()

	// dir1/dir12/file1
	// dir2/file2
	// dir3/file3
	// dir3/file4
	// file5.pdf

	dir1 := createDirectory("1", "", "dir1")
	dir12 := createDirectory("2", "1", "dir12")
	dir2 := createDirectory("3", "", "dir2")
	dir3 := createDirectory("4", "", "dir3")

	file1 := createFile("5", "2", "file1")
	file2 := createFile("6", "3", "file2")
	file3 := createFile("7", "4", "file3")
	file4 := createFile("8", "4", "file4")
	file5 := createFile("9", "", "file5")

	ctx.AddDocument(file1)
	ctx.AddDocument(file2)
	ctx.AddDocument(file3)
	ctx.AddDocument(file4)
	ctx.AddDocument(file5)
	ctx.AddDocument(dir3)
	ctx.AddDocument(dir2)
	ctx.AddDocument(dir12)
	ctx.AddDocument(dir1)

	path, _ := ctx.NodeToPath(ctx.Root())
	assert.Equal(t, "/", path)

	path, _ = ctx.NodeToPath(ctx.root.Children["1"])
	assert.Equal(t, "/dir1", path)

	path, _ = ctx.NodeToPath(ctx.root.Children["1"].Children["2"])
	assert.Equal(t, "/dir1/dir12", path)

	path, _ = ctx.NodeToPath(ctx.root.Children["1"].Children["2"].Children["5"])
	assert.Equal(t, "/dir1/dir12/file1", path)

	path, _ = ctx.NodeToPath(ctx.root.Children["3"].Children["6"])
	assert.Equal(t, "/dir2/file2", path)

	path, _ = ctx.NodeToPath(ctx.root.Children["9"])
	assert.Equal(t, "/file5", path)
}
