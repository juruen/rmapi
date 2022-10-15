package api

import (
	"errors"
	"log"
	"strings"
	"time"

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
	Auth0 struct {
		UserID string
	} `json:"auth0-profile"`
	Scopes string
	*jwt.StandardClaims
}

type SyncVersion int

const (
	Version10 SyncVersion = 10
	Version15 SyncVersion = 15
)

type UserInfo struct {
	SyncVersion SyncVersion
	User        string
}

func ParseToken(userToken string) (token *UserInfo, err error) {
	claims := UserToken{}
	jwt.ParseWithClaims(userToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), false) {
		return nil, errors.New("Token Expired")
	}

	token = &UserInfo{
		User:        claims.Auth0.UserID,
		SyncVersion: Version10,
	}

	scopes := strings.Fields(claims.Scopes)

	for _, scope := range scopes {
		switch scope {
		case "sync:fox":
			fallthrough
		case "sync:tortoise":
			fallthrough
		case "sync:hare":
			token.SyncVersion = Version15
			return
		}
	}
	return token, nil
}

// CreateApiCtx initializes an instance of ApiCtx
func CreateApiCtx(httpCtx *transport.HttpClientCtx, syncVerison SyncVersion) (ctx ApiCtx, err error) {
	switch syncVerison {
	case Version10:
		return sync10.CreateCtx(httpCtx)
	case Version15:
		return sync15.CreateCtx(httpCtx)
	default:
		log.Fatal("Unsupported sync version")
	}
	return
}
