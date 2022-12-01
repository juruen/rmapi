package sync15

import (
	"io"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	line := "hash:0:docid:0:993"
	entry, err := parseEntry(line)
	if err != nil {
		t.Error(err)
	}

	if entry.Hash != "hash" {
		t.Error("wrong hash")
	}
	if entry.DocumentID != "docid" {
		t.Error("wrong documentid")
	}

	if entry.Size != 993 {
		t.Error("wrong size")
	}
}

func TestParseIndex(t *testing.T) {
	index := `3
	0f83178c4ebe6a60fae0360b74916ee9e1faa5de1c56ab3481eccdc5cb98754f:0:fe0039fb-56a0-4561-a36f-a820f0009622.content:0:993
	17eca6c9a540c993f5f5506bb09b7a40993c02fa8f065b1a6a442e412cf2fd04:0:fe0039fb-56a0-4561-a36f-a820f0009622.metadata:0:320`
	entries, err := parseIndex(strings.NewReader(index))
	if err != nil {
		t.Error(err)
		return
	}
	if len(entries) != 2 {
		t.Error("wrong number of entries")
		return
	}
}

func TestCreateDocIndex(t *testing.T) {
	doc := &BlobDoc{
		Entry: Entry{
			Hash:       "somehash",
			DocumentID: "someid",
		},
	}
	file := &Entry{
		Hash:       "blah",
		DocumentID: "someid",
		Size:       10,
	}
	doc.AddFile(file)
	reader, err := doc.IndexReader()
	if err != nil {
		t.Error(err)
		return
	}
	index, err := io.ReadAll(reader)
	if err != nil {
		t.Error(err)
		return
	}
	expected := `3
blah:0:someid:0:10
`
	strIndex := string(index)

	if strIndex != expected {
		t.Errorf("index did not match %s", strIndex)
		return
	}
}

func TestCreateRootIndex(t *testing.T) {
	tree := HashTree{}
	doc := &BlobDoc{
		Entry: Entry{
			Hash:       "somehash",
			DocumentID: "someid"},
	}
	file := &Entry{}
	doc.AddFile(file)
	tree.Add(doc)
	reader, err := tree.IndexReader()
	if err != nil {
		t.Error(err)
		return
	}
	index, err := io.ReadAll(reader)
	if err != nil {
		t.Error(err)
		return
	}
	expected := `3
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855:80000000:someid:1:0
`
	strIndex := string(index)

	if strIndex != expected {
		t.Errorf("index did not match %s", strIndex)
		return
	}

}
