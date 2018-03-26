package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"

	"github.com/juruen/rmapi/model"
)

func DocPathToName(p string) string {
	if strings.HasSuffix(p, ".pdf") {
		return strings.TrimSuffix(path.Base(p), ".pdf")
	} else {
		return strings.TrimSuffix(path.Base(p), ".epub")
	}
}

func ToIOReader(source interface{}) (io.Reader, error) {
	var content []byte

	if source != nil {
		switch source.(type) {
		case model.DeviceTokenRequest:
			b, err := json.Marshal(source)
			if err != nil {
				return nil, err
			}

			content = b
		default:
			sources := make([]interface{}, 0)
			sources = append(sources, source)

			b, err := json.Marshal(sources)
			if err != nil {
				return nil, err
			}

			content = b
		}

	} else {
		content = make([]byte, 0)
	}

	return bytes.NewReader(content), nil
}

func CopyFile(src, dst string) (int64, error) {
	r, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer w.Close()

	n, err := io.Copy(w, r)
	if err != nil {
		return 0, err
	}

	return n, nil
}
