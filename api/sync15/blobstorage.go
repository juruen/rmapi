package sync15

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type BlobStorage struct {
	http *transport.HttpClientCtx
}

func (b *BlobStorage) PutUrl(hash string, gen int64) (string, error) {
	log.Trace.Println("fetching PUT blob url for: " + hash)
	var req model.BlobStorageRequest
	var res model.BlobStorageResponse
	req.Method = "PUT"
	req.RelativePath = hash
	if gen > 0 {
		req.Generation = strconv.FormatInt(gen, 10)
	}
	if err := b.http.Post(transport.UserBearer, config.UploadBlob, req, &res); err != nil {
		return "", err
	}
	return res.Url, nil
}

func (b *BlobStorage) GetUrl(hash string) (string, error) {
	log.Trace.Println("fetching GET blob url for: " + hash)
	var req model.BlobStorageRequest
	var res model.BlobStorageResponse
	req.Method = "GET"
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
	url, err := b.PutUrl(hash, -1)
	if err != nil {
		return err
	}
	log.Trace.Println("put url: " + url)

	_, err = b.http.PutBlobStream(url, -1, reader)
	return err
}

func (b *BlobStorage) SyncComplete() error {
	return b.http.Post(transport.UserBearer, config.SyncComplete, nil, nil)
}

func (b *BlobStorage) WriteRootIndex(roothash string, gen int64) (int64, error) {
	log.Info.Println("writing root with gen: ", gen)
	url, err := b.PutUrl("root", gen)
	if err != nil {
		return 0, err
	}
	log.Trace.Println("got root url:", url)
	reader := bytes.NewBufferString(roothash)

	gen, err = b.http.PutBlobStream(url, gen, reader)
	return gen, err

}
func (b *BlobStorage) GetRootIndex() (string, int64, error) {

	url, err := b.GetUrl("root")
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
	content, err := ioutil.ReadAll(blob)
	if err != nil {
		return "", 0, err
	}
	log.Info.Println("got root gen:", gen)
	return string(content), gen, nil

}
