package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile    = ".rmapi"
	defaultConfigFileXDG = "rmapi.conf"
	appName              = "rmapi"
	configFileEnvVar     = "RMAPI_CONFIG"
)

/*
ConfigPath returns the path to the config file. It will check the following in order:
  - If the RMAPI_CONFIG environment variable is set, it will use that path.
  - If a config file exists in the user's home dir as described by os.UserHomeDir, it will use that.
  - Otherwise, it will use the XDG config dir, as described by os.UserConfigDir.
*/
func ConfigPath() (string, error) {
	if config, ok := os.LookupEnv(configFileEnvVar); ok {
		return config, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	config := filepath.Join(home, defaultConfigFile)

	//return config in home if exists
	if _, err := os.Stat(config); err == nil {
		return config, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Warning.Println("cannot determine config dir, using HOME", err)
		return config, nil
	}

	xdgConfigDir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(xdgConfigDir, 0700); err != nil {
		log.Error.Panicln("cannot create config dir "+xdgConfigDir, err)
	}
	config = filepath.Join(xdgConfigDir, defaultConfigFileXDG)

	return config, nil

}

func LoadTokens(path string) model.AuthTokens {
	tokens := model.AuthTokens{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Trace.Printf("config fail %s doesn't exist/n", path)
		return tokens
	}

	content, err := ioutil.ReadFile(path)

	if err != nil {
		log.Warning.Printf("failed to open %s with %s/n", path, err)
		return tokens
	}

	err = yaml.Unmarshal(content, &tokens)

	if err != nil {
		log.Error.Fatalln("failed to parse", path)
	}

	return tokens
}

func SaveTokens(path string, tokens model.AuthTokens) {
	content, err := yaml.Marshal(tokens)

	if err != nil {
		log.Warning.Println("failed to marsha tokens", err)
	}

	ioutil.WriteFile(path, content, 0600)

	if err != nil {
		log.Warning.Println("failed to save config to", path)
	}
}
