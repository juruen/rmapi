package fusefs

import (
	"fmt"
	"os"
	"time"

	"github.com/peerdavid/rmapi/config"
	"github.com/peerdavid/rmapi/log"
	"github.com/peerdavid/rmapi/model"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/peerdavid/rmapi/api"
)

type fuseFs struct {
	apiCtx    *api.ApiCtx
	root      *fuseNode
	backedDir string
}

func NewFuseFsRoot(ctx *api.ApiCtx) nodefs.Node {

	backedDirPath := config.ConfigPath() + "-cache"

	os.Mkdir(backedDirPath, 0700)

	fs := &fuseFs{
		apiCtx:    ctx,
		backedDir: backedDirPath,
	}

	fs.root = fs.newNode("")

	return fs.root
}

type rmapiNode interface {
	nodefs.Node
	NodeId() string
}

type fuseNode struct {
	rmapiNode
	Fs    *fuseFs
	Id    string
	inode *nodefs.Inode
}

func (n *fuseNode) NodeId() string {
	return n.Id
}

func (fs *fuseFs) newNode(id string) *fuseNode {
	return &fuseNode{Fs: fs, Id: id}
}

func (fs *fuseFs) Root() nodefs.Node {
	return fs.root
}

func (n *fuseNode) StatFs() *fuse.StatfsOut {
	return &fuse.StatfsOut{}
}

func (n *fuseNode) Deletable() bool {
	return false
}

func (n *fuseNode) Lookup(out *fuse.Attr, name string, context *fuse.Context) (node *nodefs.Inode, code fuse.Status) {
	cNode := n.modelNode()
	log.Trace.Println("Lookup", name, "on", cNode.Name())

	tNode, err := cNode.FindByName(name)

	if err != nil {
		return nil, fuse.ENOENT
	}

	out.Mode = nodeMode(tNode)

	ch := n.Fs.newNode(tNode.Id())
	n.Inode().NewChild(name, true, ch)

	return ch.Inode(), fuse.OK
}

func (n *fuseNode) Mkdir(name string, mode uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	log.Trace.Println("mkdir", name, "on", n.Id)

	document, err := n.Fs.apiCtx.CreateDir(n.Id, name)

	if err != nil {
		log.Error.Println("failed to create directory", err)
		return nil, fuse.EIO
	}

	n.Fs.apiCtx.Filetree.AddDocument(document)

	ch := n.Fs.newNode(document.ID)
	n.Inode().NewChild(name, true, ch)

	return ch.Inode(), fuse.OK
}

func (n *fuseNode) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	log.Trace.Println("unlink", name, "on", n.Id)

	node, err := n.modelNode().FindByName(name)

	if err != nil {
		log.Error.Println("entry", n.Id, "doesn't exist")
		return fuse.ENOENT
	}

	if n.Fs.apiCtx.DeleteEntry(node) != nil {
		log.Error.Println("failed to delete entry", name, "on", n.Id, err)
		return fuse.EBUSY
	}

	n.Fs.apiCtx.Filetree.DeleteNode(node)

	return fuse.OK
}

func (n *fuseNode) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	return n.Unlink(name, context)
}

func (n *fuseNode) OpenDir(context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	s := make([]fuse.DirEntry, 0)

	cNode := n.modelNode()

	if cNode == nil {
		return nil, fuse.ENOENT
	}

	for _, node := range cNode.Children {
		s = append(s, fuse.DirEntry{Name: node.Name(), Mode: nodeMode(node)})
	}

	return s, fuse.OK
}

func (n *fuseNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	cNode := n.modelNode()

	log.Trace.Println("GetAttr (fuseNode) file", file, "name", cNode.Name())

	if file != nil {
		return file.GetAttr(out)
	}

	stat, err := os.Stat(n.Fs.backedDir + "/" + cNode.Id())

	if err == nil {
		out.Size = uint64(stat.Size())
	}

	out.Mode = nodeMode(n.modelNode())

	return fuse.OK
}

func (n *fuseNode) Rename(oldName string, newParent nodefs.Node, newName string, context *fuse.Context) (code fuse.Status) {
	log.Trace.Println("Rename ", oldName, "from", n.modelNode().Name(), "to", newName)

	srcParent := n.modelNode()

	if srcParent == nil {
		log.Error.Println("source parent entry doesn't exist any more")
		return fuse.EBADF
	}

	src, err := srcParent.FindByName(oldName)

	if err != nil {
		log.Error.Println("source entry doesn't exist")
		return fuse.ENOENT
	}

	newParentFuseNode, ok := newParent.(rmapiNode)

	if !ok {
		log.Error.Fatalln("failed to cast to rmapiNode")
		return fuse.EBADF
	}

	dstParent := n.Fs.apiCtx.Filetree.NodeById(newParentFuseNode.NodeId())

	if dstParent == nil {
		log.Error.Println("destination parent entry doesn't exist any more")
		return fuse.ENOENT
	}

	newNode, err := n.Fs.apiCtx.MoveEntry(src, dstParent, newName)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return fuse.EIO
	}

	n.Fs.apiCtx.Filetree.MoveNode(src, newNode)

	return fuse.OK
}

func (n *fuseNode) modelNode() *model.Node {
	return n.Fs.apiCtx.Filetree.NodeById(n.Id)
}

func nodeMode(n *model.Node) uint32 {
	if n.IsDirectory() {
		return fuse.S_IFDIR | 0755
	} else {
		return fuse.S_IFREG | 0644
	}
}

func (fs *fuseNode) OnUnmount() {
}

func (fs *fuseNode) OnMount(conn *nodefs.FileSystemConnector) {
}

func (n *fuseNode) SetInode(node *nodefs.Inode) {
	n.inode = node
}

func (n *fuseNode) Inode() *nodefs.Inode {
	return n.inode
}

func (n *fuseNode) OnForget() {
}

func (n *fuseNode) Access(mode uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Readlink(c *fuse.Context) ([]byte, fuse.Status) {
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Symlink(name string, content string, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Link(name string, existing nodefs.Node, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, newNode *nodefs.Inode, code fuse.Status) {
	return nil, nil, fuse.ENOSYS
}

func (n *fuseNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	cNode := n.modelNode()
	log.Trace.Println("Open ", cNode.Name())

	backedFile := fmt.Sprintf("%s/%s", n.Fs.backedDir, cNode.Id())

	err := n.Fs.apiCtx.FetchDocument(cNode.Id(), backedFile)

	if err != nil {
		log.Error.Println("Failed to fetch", cNode.Name(), cNode.Id(), err)
		return nil, fuse.EIO
	}

	f, err := os.Open(backedFile)
	if err != nil {
		log.Error.Println("Failed to open", cNode.Name(), cNode.Id(), err)
		return nil, fuse.EIO
	}

	return NewFuseFile(f), fuse.OK
}

func (n *fuseNode) Flush(file nodefs.File, openFlags uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) GetXAttr(attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	return nil, fuse.ENOATTR
}

func (n *fuseNode) RemoveXAttr(attr string, context *fuse.Context) fuse.Status {
	return fuse.ENOSYS
}

func (n *fuseNode) SetXAttr(attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	return fuse.ENOSYS
}

func (n *fuseNode) ListXAttr(context *fuse.Context) (attrs []string, code fuse.Status) {
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Chmod(file nodefs.File, perms uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Chown(file nodefs.File, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Utimens(file nodefs.File, atime *time.Time, mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Fallocate(file nodefs.File, off uint64, size uint64, mode uint32, context *fuse.Context) (code fuse.Status) {
	return fuse.ENOSYS
}

func (n *fuseNode) Read(file nodefs.File, dest []byte, off int64, context *fuse.Context) (fuse.ReadResult, fuse.Status) {
	if file != nil {
		return file.Read(dest, off)
	}
	return nil, fuse.ENOSYS
}

func (n *fuseNode) Write(file nodefs.File, data []byte, off int64, context *fuse.Context) (written uint32, code fuse.Status) {
	return 0, fuse.ENOSYS
}
