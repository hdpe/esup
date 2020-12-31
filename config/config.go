package config

import (
	"fmt"
	viperlib "github.com/spf13/viper"
	"strings"
)

func NewConfig() (Config, error) {
	viper := viperlib.New()
	viper.SetDefault("server.address", "http://localhost:9200")
	viper.SetDefault("changelog.index", "esup-changelog0")
	viper.SetDefault("pipelines.directory", "./pipelines")
	viper.SetDefault("indexSets.directory", "./indexSets")
	viper.SetDefault("documents.directory", "./documents")
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
			Address: viper.GetString("server.address"),
			ApiKey:  viper.GetString("server.apiKey"),
		},
		PrototypeConfig{Environment: viper.GetString("prototype.environment")},
		ChangelogConfig{Index: viper.GetString("changelog.index")},
		IndexSetsConfig{Directory: viper.GetString("indexSets.directory")},
		PipelinesConfig{Directory: viper.GetString("pipelines.directory")},
		DocumentsConfig{Directory: viper.GetString("documents.directory")},
		PreprocessConfig{IncludesDirectory: viper.GetString("preprocess.includesDirectory")},
	}, nil
}

type Config struct {
	Server     ServerConfig
	Prototype  PrototypeConfig
	Changelog  ChangelogConfig
	IndexSets  IndexSetsConfig
	Pipelines  PipelinesConfig
	Documents  DocumentsConfig
	Preprocess PreprocessConfig
}

type ServerConfig struct {
	Address string
	ApiKey  string
}

type PrototypeConfig struct {
	Environment string
}

type ChangelogConfig struct {
	Index string
}

type IndexSetsConfig struct {
	Directory string
}

type PipelinesConfig struct {
	Directory string
}

type DocumentsConfig struct {
	Directory string
}

type PreprocessConfig struct {
	IncludesDirectory string
}
