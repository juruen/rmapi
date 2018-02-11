package util

import (
	"bytes"
	"encoding/json"
	"io"
	"path"
	"strings"

	"github.com/juruen/rmapi/common"
)

func PdfPathToName(pdfPath string) string {
	return strings.TrimSuffix(path.Base(pdfPath), ".pdf")
}

func ToIOReader(source interface{}) (io.Reader, error) {
	var content []byte

	if source != nil {
		switch source.(type) {
		case common.DeviceTokenRequest:
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
