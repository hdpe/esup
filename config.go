package main

import (
	viperlib "github.com/spf13/viper"
	"strings"
)

func newConfig() Config {
	viper := viperlib.New()
	viper.SetDefault("server.address", "http://localhost:9200")
	viper.SetDefault("changelog.index", "esup-changelog0")
	viper.SetDefault("pipelines.directory", "./pipelines")
	viper.SetDefault("indexSets.directory", "./indexSets")
	viper.SetDefault("preprocess.includesDirectory", "./includes")

	viper.SetConfigType("yaml")
	viper.SetConfigFile("esup.config")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	return Config{
		ServerConfig{
			address: viper.GetString("server.address"),
		},
		ChangelogConfig{index: viper.GetString("changelog.index")},
		IndexSetsConfig{directory: viper.GetString("indexSets.directory")},
		PipelinesConfig{directory: viper.GetString("pipelines.directory")},
		PreprocessConfig{includesDirectory: viper.GetString("preprocess.includesDirectory")},
	}
}

type Config struct {
	server     ServerConfig
	changelog  ChangelogConfig
	indexSets  IndexSetsConfig
	pipelines  PipelinesConfig
	preprocess PreprocessConfig
}

type ServerConfig struct {
	address string
	apiKey  string
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
