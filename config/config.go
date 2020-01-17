// App-level config
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type GithubOAuthConf struct {
	ClientID     string `yaml:"ClientID"`
	ClientSecret string `yaml:"ClientSecret"`
}

var (
	GithubOAuth GithubOAuthConf
	JWTSecret   string
	AppRoot     string
)

func init() {
	configBytes, err := ioutil.ReadFile("./conf/config.yaml")
	if err != nil {
		panic(err)
	}
	type appConfig struct {
		GithubOAuth GithubOAuthConf `yaml:"GithubOAuth"`
		JWTSecret   string          `yaml:"JWTSecret"`
		AppRoot     string          `yaml:"AppRoot"`
	}
	var appConf appConfig
	err = yaml.Unmarshal(configBytes, &appConf)
	if err != nil {
		panic(err)
	}

	GithubOAuth = appConf.GithubOAuth
	JWTSecret = appConf.JWTSecret
	AppRoot = appConf.AppRoot
}
