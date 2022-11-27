package sync15

import "io"

type RemoteStorage interface {
	GetRootIndex() (hash string, generation int64, err error)
	GetReader(hash string) (io.ReadCloser, error)
	Concurrent() int
}

type RemoteStorageWriter interface {
	UpdateRootIndex(hash string, generation int64) (gen int64, err error)
	GetWriter(hash string, writer io.WriteCloser) error
}
