package config

import (
	"io/ioutil"
	"os"
	"os/user"
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

func ConfigPath() (config string) {
	configFile, ok := os.LookupEnv(configFileEnvVar)
	if ok {
		return configFile
	}

	user, err := user.Current()
	if err != nil {
		log.Error.Panicln("failed to get current user:", err)
	}

	home := user.HomeDir
	config = filepath.Join(home, defaultConfigFile)

	//return config in home if exists
	if _, err := os.Stat(config); err == nil {
		return
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Warning.Println("cannot determine config dir, using HOME", err)
		return
	}

	xdgConfigDir := filepath.Join(configDir, appName)
	err = os.MkdirAll(xdgConfigDir, 0700)
	if err != nil {
		log.Error.Panicln("cannot create config dir "+xdgConfigDir, err)
	}
	config = filepath.Join(xdgConfigDir, defaultConfigFileXDG)

	return

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
