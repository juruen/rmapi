package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = ".rmapi"
)

func configPath() string {
	home := os.Getenv("HOME")

	return fmt.Sprintf("%s/%s", home, defaultConfigFile)
}

func loadTokens(path string) AuthTokens {
	tokens := AuthTokens{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		Trace.Printf("config fail %s doesn't exist", path)
		return tokens
	}

	content, err := ioutil.ReadFile(path)

	if err != nil {
		Warning.Println("failed to open %s with %s", path, err)
		return tokens
	}

	err = yaml.Unmarshal(content, &tokens)

	if err != nil {
		Error.Fatalln("failed to parse %s", path)
	}

	return tokens
}

func saveTokens(path string, tokens AuthTokens) {
	content, err := yaml.Marshal(tokens)

	if err != nil {
		Warning.Println("failed to marsha tokens %s", err)
	}

	ioutil.WriteFile(path, content, 0600)

	if err != nil {
		Warning.Println("failed to save config to %s", path)
	}
}
