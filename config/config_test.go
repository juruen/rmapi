package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/juruen/rmapi/model"
	"github.com/stretchr/testify/assert"
)

func TestSaveLoadConfig(t *testing.T) {
	tokens := model.AuthTokens{
		DeviceToken: "foo",
		UserToken:   "bar",
	}

	f, err := ioutil.TempFile("", "rmapitmp")

	if err != nil {
		panic(fmt.Sprintln("failed to create temp file"))
	}

	path := f.Name()

	defer os.Remove(path)

	SaveTokens(path, tokens)

	savedTokens := LoadTokens(path)

	assert.Equal(t, "foo", savedTokens.DeviceToken)
	assert.Equal(t, "bar", savedTokens.UserToken)
}

func TestConfigPath(t *testing.T) {
	// let's not mess with the user's home dir
	home := "HOME"
	switch runtime.GOOS {
	case "windows":
		home = "USERPROFILE"
	case "plan9":
		home = "home"
	}
	if err := os.Setenv(home, os.TempDir()); err != nil {
		t.Error(err)
	}

	tearDown := func() {
		_ = os.Unsetenv(configFileEnvVar)
		_ = os.Remove(filepath.Join(os.TempDir(), defaultConfigFile))
	}

	tests := []struct {
		name    string
		setup   func()
		want    string
		wantErr bool
	}{
		{
			name:  "no home no env config exists",
			setup: func() {},
			want: func() string {
				xdgConfigDir, err := os.UserConfigDir()
				if err != nil {
					t.Error(err)
				}
				xdgConfig := filepath.Join(xdgConfigDir, appName, defaultConfigFileXDG)
				return xdgConfig
			}(),
		},
		{
			name: "home config exists",
			setup: func() {
				homeConfig := filepath.Join(os.TempDir(), defaultConfigFile)
				if err := ioutil.WriteFile(homeConfig, []byte("test"), 0644); err != nil {
					t.Error(err)
				}
			},
			want: filepath.Join(os.TempDir(), defaultConfigFile),
		},
		{
			name: "env config exists",
			setup: func() {
				if err := os.Setenv(configFileEnvVar, filepath.Join(os.TempDir(), "rmapi.yaml")); err != nil {
					t.Error(err)
				}
			},
			want: filepath.Join(os.TempDir(), "rmapi.yaml"),
		},
	}

	// Can't allow parallel execution because of shared file state
	wg := sync.WaitGroup{}
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			defer wg.Done()
			defer tearDown()
			tt.setup()

			got, err := ConfigPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfigPath() = %v, want %v", got, tt.want)
			}
		})
		wg.Wait()
	}
}
