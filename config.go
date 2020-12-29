package main

import (
	"fmt"
	viperlib "github.com/spf13/viper"
	"strings"
)

func newConfig() (Config, error) {
	viper := viperlib.New()
	viper.SetDefault("server.address", "http://localhost:9200")
	viper.SetDefault("changelog.index", "esup-changelog0")
	viper.SetDefault("pipelines.directory", "./pipelines")
	viper.SetDefault("indexSets.directory", "./indexSets")
	viper.SetDefault("preprocess.includesDirectory", "./includes")

	viper.SetConfigType("yml")
	viper.SetConfigName("esup.config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viperlib.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("couldn't read esup.config.yml: %w", err)
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	return Config{
		ServerConfig{
			address: viper.GetString("server.address"),
			apiKey:  viper.GetString("server.apiKey"),
		},
		PrototypeConfig{environment: viper.GetString("prototype.environment")},
		ChangelogConfig{index: viper.GetString("changelog.index")},
		IndexSetsConfig{directory: viper.GetString("indexSets.directory")},
		PipelinesConfig{directory: viper.GetString("pipelines.directory")},
		PreprocessConfig{includesDirectory: viper.GetString("preprocess.includesDirectory")},
	}, nil
}

type Config struct {
	server     ServerConfig
	prototype  PrototypeConfig
	changelog  ChangelogConfig
	indexSets  IndexSetsConfig
	pipelines  PipelinesConfig
	preprocess PreprocessConfig
}

type ServerConfig struct {
	address string
	apiKey  string
}

type PrototypeConfig struct {
	environment string
}

type ChangelogConfig struct {
	index string
}

type IndexSetsConfig struct {
	directory string
}

type PipelinesConfig struct {
	directory string
}

type PreprocessConfig struct {
	includesDirectory string
}
