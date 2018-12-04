package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = ".rmapi"
)

func ConfigPath() string {
	user, err := user.Current()
	if err != nil {
		log.Error.Panicln("failed to get current user:", err)
	}

	home := user.HomeDir

	return fmt.Sprintf("%s/%s", home, defaultConfigFile)
}

func LoadTokens(path string) model.AuthTokens {
	tokens := model.AuthTokens{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Trace.Printf("config fail %s doesn't exist", path)
		return tokens
	}

	content, err := ioutil.ReadFile(path)

	if err != nil {
		log.Warning.Println("failed to open %s with %s", path, err)
		return tokens
	}

	err = yaml.Unmarshal(content, &tokens)

	if err != nil {
		log.Error.Fatalln("failed to parse %s", path)
	}

	return tokens
}

func SaveTokens(path string, tokens model.AuthTokens) {
	content, err := yaml.Marshal(tokens)

	if err != nil {
		log.Warning.Println("failed to marsha tokens %s", err)
	}

	ioutil.WriteFile(path, content, 0600)

	if err != nil {
		log.Warning.Println("failed to save config to %s", path)
	}
}
