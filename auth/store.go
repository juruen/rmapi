package auth

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const defaultFile = ".rmapi"

// TokenSet contains tokens needed for the Remarkable Cloud authentication.
type TokenSet struct {
	// DeviceToken is a token that gets returned after
	// registering a device to the API. It can be fetched using the RegisterDevice method
	// or can be set manually for caching purposes.
	DeviceToken string `yaml:"devicetoken"`

	// UserToken is a token that gets returned as a second step, by the help of a previously
	// fetched DeviceToken and is actually used to make the proper authenticated
	// HTTP calls to the Remarkable API. Set to empty to force fetching it again.
	UserToken string `yaml:"usertoken"`
}

// TokenStore is an interface that will allow
// to load and save tokens needed for the Remarkable Cloud API.
type TokenStore interface {
	Save(t TokenSet) error
	Load() (TokenSet, error)
}

// FileTokenStore implements TokenStore by fetching and saving
// tokens to a plain file.
type FileTokenStore struct {
	Path string
}

// path returns the path of the file containing
// the configuration. If path is not defined, it falls back
// to the default one ($HOME/.rmapi).
func (ft *FileTokenStore) path() string {
	if ft.Path != "" {
		return ft.Path
	}

	// assume not returning error
	usr, _ := user.Current()

	return filepath.Join(usr.HomeDir, defaultFile)
}

// Save will persist a TokenSet into a yaml file.
func (ft *FileTokenStore) Save(t TokenSet) error {
	content, err := yaml.Marshal(t)

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(ft.path(), content, 0600); err != nil {
		return err
	}

	return nil
}

// Load will return a TokenSet with content populated
// from a yaml file containing the values.
func (ft *FileTokenStore) Load() (TokenSet, error) {
	// return empty struct if file does not exist
	if _, err := os.Stat(ft.path()); os.IsNotExist(err) {
		return TokenSet{}, nil
	}

	content, err := ioutil.ReadFile(ft.path())
	if err != nil {
		return TokenSet{}, err
	}

	var tks TokenSet
	err = yaml.Unmarshal(content, &tks)
	if err != nil {
		return TokenSet{}, err
	}

	return tks, nil
}
