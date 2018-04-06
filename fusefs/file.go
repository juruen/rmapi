package fusefs

import (
	"os"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/juruen/rmapi/log"
)

type fuseFile struct {
	localFile *os.File
	inode     *nodefs.Inode
}

// NewfuseFile returns a File instance that returns ENOSYS for
// every operation.
func NewFuseFile(f *os.File) nodefs.File {
	return &fuseFile{f, nil}
}

func (f *fuseFile) SetInode(inode *nodefs.Inode) {
	f.inode = inode
}

func (f *fuseFile) InnerFile() nodefs.File {
	return nil
}

func (f *fuseFile) String() string {
	return f.localFile.Name()
}

func (f *fuseFile) Read(buf []byte, off int64) (fuse.ReadResult, fuse.Status) {
	log.Trace.Println("Read len(buf)", len(buf), "off", off)

	return fuse.ReadResultFd(f.localFile.Fd(), off, len(buf)), fuse.OK
}

func (f *fuseFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	return 0, fuse.ENOSYS
}

func (f *fuseFile) Flock(flags int) fuse.Status {
	return fuse.ENOSYS
}

func (f *fuseFile) Flush() fuse.Status {
	log.Trace.Println("Flush", f.localFile.Name())
	f.localFile.Close()
	return fuse.OK
}

func (f *fuseFile) Release() {
}

func (f *fuseFile) GetAttr(attr *fuse.Attr) fuse.Status {
	log.Trace.Println("GetAttr", f.localFile.Name())

	stat, err := f.localFile.Stat()

	if err != nil {
		log.Error.Println("failed to fetch stat from", f.localFile.Name())
		return fuse.EBADF
	}

	attr.Size = uint64(stat.Size())

	return fuse.OK
}

func (f *fuseFile) Fsync(flags int) (code fuse.Status) {
	return fuse.ENOSYS
}

func (f *fuseFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {
	return fuse.ENOSYS
}

func (f *fuseFile) Truncate(size uint64) fuse.Status {
	return fuse.ENOSYS
}

func (f *fuseFile) Chown(uid uint32, gid uint32) fuse.Status {
	return fuse.ENOSYS
}

func (f *fuseFile) Chmod(perms uint32) fuse.Status {
	return fuse.ENOSYS
}

func (f *fuseFile) Allocate(off uint64, size uint64, mode uint32) (code fuse.Status) {
	return fuse.ENOSYS
}
