package archive

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriter(t *testing.T) {
	zipFile, err := os.Create("test_writer.zip")
	if err != nil {
		t.Error(err)
	}
	defer zipFile.Close()

	archive := NewWriter(zipFile, "384327f5-133e-49c8-82ff-30aa19f3cfa4")
	defer archive.Close()

	// Content
	contentWriter, err := archive.CreateContent()
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(contentWriter, strings.NewReader("content"))
	if err != nil {
		t.Error(err)
	}

	// Pagedata
	pageDataWriter, err := archive.CreatePagedata()
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(pageDataWriter, strings.NewReader("pagedata"))
	if err != nil {
		t.Error(err)
	}

	// Thumbnail
	thumbnailWriter, err := archive.CreateThumbnail(0)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(thumbnailWriter, strings.NewReader("thumbnail"))
	if err != nil {
		t.Error(err)
	}

	// Page
	pageWriter, err := archive.CreatePage(0)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(pageWriter, strings.NewReader("page data"))
	if err != nil {
		t.Error(err)
	}

	// Page metadata
	pageMetadataWriter, err := archive.CreatePageMetadata(0)
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(pageMetadataWriter, strings.NewReader("page metadata"))
	if err != nil {
		t.Error(err)
	}
}
