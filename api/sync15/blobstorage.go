package sync15

import (
	"io"
	"io/ioutil"

	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type BlobStorage struct {
	http *transport.HttpClientCtx
}

func (b *BlobStorage) GetUrl(method, hash string) (string, error) {
	var req model.BlobStorageRequest
	var res model.BlobStorageResponse
	req.Method = method
	req.RelativePath = hash
	if err := b.http.Post(transport.UserBearer, config.DownloadBlob, req, &res); err != nil {
		return "", err
	}
	return res.Url, nil
}

func (b *BlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	log.Info.Println("get blobl: " + hash)
	url, err := b.GetUrl("GET", hash)
	if err != nil {
		return nil, err
	}

	blob, _, err := b.http.GetBlobStream(url)
	return blob, err
}
func (b *BlobStorage) GetRootIndex() (string, int64, error) {

	log.Info.Println("get root")
	url, err := b.GetUrl("GET", "root")
	if err != nil {
		return "", 0, err
	}
	log.Info.Println("got url:", url)
	blob, gen, err := b.http.GetBlobStream(url)
	if err != nil {
		return "", 0, err
	}
	content, err := ioutil.ReadAll(blob)
	if err != nil {
		return "", 0, err
	}
	log.Info.Println("got gen:", gen)
	return string(content), gen, nil

}
