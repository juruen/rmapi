package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/juruen/rmapi/api/sync10"
	"github.com/juruen/rmapi/api/sync15"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type ApiCtx interface {
	Filetree() *filetree.FileTreeCtx
	FetchDocument(docId, dstPath string) error
	CreateDir(parentId, name string) (*model.Document, error)
	UploadDocument(parentId string, sourceDocPath string, notify bool) (*model.Document, error)
	MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error)
	DeleteEntry(node *model.Node) error
	SyncComplete() error
	Nuke() error
}

type UserToken struct {
	Scopes string
	*jwt.StandardClaims
}

// CreateApiCtx initializes an instance of ApiCtx
func CreateApiCtx(http *transport.HttpClientCtx) (ctx ApiCtx, err error) {
	userToken := http.Tokens.UserToken
	claims := UserToken{}
	jwt.ParseWithClaims(userToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		return
	}
	fld := strings.Fields(claims.Scopes)
	isSync15 := false

forloop:
	for _, f := range fld {
		switch f {
		case "sync:fox":
			fallthrough
		case "sync:tortoise":
			fallthrough
		case "sync:hare":
			isSync15 = true
			break forloop
		}
	}
	if isSync15 {
		fmt.Fprintln(os.Stderr, `WARNING!!! Using the new 1.5 sync, this has not been fully tested yet!!! first sync can be slow`)
		ctx, err = sync15.CreateCtx(http)
	} else {
		ctx, err = sync10.CreateCtx(http)
	}
	return
}
