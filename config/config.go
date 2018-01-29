package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/common"
	"github.com/juruen/rmapi/log"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = ".rmapi"
)

func ConfigPath() string {
	home := os.Getenv("HOME")

	return fmt.Sprintf("%s/%s", home, defaultConfigFile)
}

func LoadTokens(path string) common.AuthTokens {
	tokens := common.AuthTokens{}

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

func SaveTokens(path string, tokens common.AuthTokens) {
	content, err := yaml.Marshal(tokens)

	if err != nil {
		log.Warning.Println("failed to marsha tokens %s", err)
	}

	ioutil.WriteFile(path, content, 0600)

	if err != nil {
		log.Warning.Println("failed to save config to %s", path)
	}
}
