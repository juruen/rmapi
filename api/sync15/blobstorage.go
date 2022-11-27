package sync15

import (
	"bytes"
	"io"
	"net/http"

	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type BlobStorage struct {
	http        *transport.HttpClientCtx
	concurrency int
}

func NewBlobStorage(http *transport.HttpClientCtx) *BlobStorage {
	return &BlobStorage{
		http: http,
	}
}

const ROOT_NAME = "root"

func (b *BlobStorage) PutRootUrl(hash string, gen int64) (string, int64, error) {
	log.Trace.Println("fetching  ROOT url for: " + hash)
	req := model.BlobRootStorageRequest{
		Method:       http.MethodPut,
		RelativePath: ROOT_NAME,
		RootSchema:   hash,
		Generation:   gen,
	}
	var res model.BlobStorageResponse

	if err := b.http.Post(transport.UserBearer, config.UploadBlob, req, &res); err != nil {
		return "", 0, err
	}
	return res.Url, res.MaxUploadSizeBytes, nil
}
func (b *BlobStorage) PutUrl(hash string) (string, int64, error) {
	log.Trace.Println("fetching PUT blob url for: " + hash)
	var req model.BlobStorageRequest
	var res model.BlobStorageResponse
	req.Method = http.MethodPut
	req.RelativePath = hash
	if err := b.http.Post(transport.UserBearer, config.UploadBlob, req, &res); err != nil {
		return "", 0, err
	}
	return res.Url, res.MaxUploadSizeBytes, nil
}

func (b *BlobStorage) GetUrl(hash string) (string, error) {
	log.Trace.Println("fetching GET blob url for: " + hash)
	var req model.BlobStorageRequest
	var res model.BlobStorageResponse
	req.Method = http.MethodGet
	req.RelativePath = hash
	if err := b.http.Post(transport.UserBearer, config.DownloadBlob, req, &res); err != nil {
		return "", err
	}
	return res.Url, nil
}

func (b *BlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	url, err := b.GetUrl(hash)
	if err != nil {
		return nil, err
	}
	log.Trace.Println("get url: " + url)

	blob, _, err := b.http.GetBlobStream(url)
	return blob, err
}

func (b *BlobStorage) UploadBlob(hash string, reader io.Reader) error {
	url, size, err := b.PutUrl(hash)
	if err != nil {
		return err
	}
	log.Trace.Println("put url: " + url)

	return b.http.PutBlobStream(url, reader, size)
}

// SyncComplete notifies that the sync is done
func (b *BlobStorage) SyncComplete(gen int64) error {
	req := model.SyncCompletedRequest{
		Generation: gen,
	}
	return b.http.Post(transport.UserBearer, config.SyncComplete, req, nil)
}

func (b *BlobStorage) WriteRootIndex(roothash string, gen int64) (int64, error) {
	log.Info.Println("writing root with gen: ", gen)
	url, maxRequestSize, err := b.PutRootUrl(roothash, gen)
	if err != nil {
		return 0, err
	}
	log.Trace.Println("got root url:", url)
	reader := bytes.NewBufferString(roothash)

	return b.http.PutRootBlobStream(url, gen, maxRequestSize, reader)
}
func (b *BlobStorage) GetRootIndex() (string, int64, error) {
	url, err := b.GetUrl(ROOT_NAME)
	if err != nil {
		return "", 0, err
	}
	log.Info.Println("got root get url:", url)
	blob, gen, err := b.http.GetBlobStream(url)
	if err == transport.ErrNotFound {
		return "", 0, nil

	}
	if err != nil {
		return "", 0, err
	}
	content, err := io.ReadAll(blob)
	if err != nil {
		return "", 0, err
	}
	log.Info.Println("got root gen:", gen)
	return string(content), gen, nil

}
