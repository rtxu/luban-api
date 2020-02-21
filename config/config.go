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

type MysqlConf struct {
	Host     string `yaml:"Host"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	Database string `yaml:"Database"`
}

type AppConfig struct {
	GithubOAuth GithubOAuthConf `yaml:"GithubOAuth"`
	JWTSecret   string          `yaml:"JWTSecret"`
	AppRoot     string          `yaml:"AppRoot"`
	Mysql       MysqlConf       `yaml:"Mysql"`
}

func LoadConfig() (AppConfig, error) {
	var appConf AppConfig
	configBytes, err := ioutil.ReadFile("./conf/config.yaml")
	if err != nil {
		return appConf, err
	}
	err = yaml.Unmarshal(configBytes, &appConf)
	if err != nil {
		return appConf, err
	}

	return appConf, nil
}
