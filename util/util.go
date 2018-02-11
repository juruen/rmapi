package util

import (
	"path"
	"strings"
)

func PdfPathToName(pdfPath string) string {
	return strings.TrimSuffix(path.Base(pdfPath), ".pdf")
}
