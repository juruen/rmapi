// +build withauth

package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/juruen/rmapi/auth"
)

const testUUID = "883ba04f-606c-41b7-8903-8d113356850f"
const testName = "test-api"

func TestList(t *testing.T) {
	cli := NewClient(auth.New().Client())

	docs, err := cli.List()
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	for _, doc := range docs {
		t.Log(doc)
	}
}

func TestUpload(t *testing.T) {
	cli := NewClient(auth.New().Client())

	// open test file
	file, err := os.Open("test.zip")
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	defer file.Close()

	if err := cli.Upload(testUUID, testName, file); err != nil {
		t.Fatalf("test: %v", err)
	}
}

func TestDownload(t *testing.T) {
	cli := NewClient(auth.New().Client())

	file, err := ioutil.TempFile("", "rmapi-test-*.zip")
	if err != nil {
		t.Fatalf("test: can't create temporary file: %v", err)
	}
	t.Log("path of the created file", file.Name())

	if err := cli.Download(testUUID, file); err != nil {
		t.Fatalf("test: can't download file: %v", err)
	}
}

func TestCreateFolder(t *testing.T) {
	cli := NewClient(auth.New().Client())

	if _, err := cli.CreateFolder("test-folder", ""); err != nil {
		t.Fatalf("test: can't create folder: %v", err)
	}
}

func TestBookmark(t *testing.T) {
	cli := NewClient(auth.New().Client())

	doc := Document{
		ID:         testUUID,
		Name:       testName,
		Bookmarked: true,
	}

	if err := cli.Metadata(doc); err != nil {
		t.Fatalf("test: can't bookmark document: %v", err)
	}
}

func TestDelete(t *testing.T) {
	cli := NewClient(auth.New().Client())

	if err := cli.Delete(testUUID); err != nil {
		t.Fatalf("test: can't delete document: %v", err)
	}
}
