package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// type Token struct {
// 	Name     string
// 	Contract string
// }

type General struct {
	Mode     string `yaml:"mode"`
	logLevel string `yaml:"logLevel"`
}

type Blockchain struct {
	Network string `yaml:"network"`
}

type Pair struct {
	Token0 string `yaml:"token0"`
	Token1 string `yaml:"token1"`
	Pool   string `yaml:"pool"`
}

type Dex struct {
	Name     string `yaml:"name"`
	Contract string `yaml:"contractAddress"`
	Pairs    []Pair `yaml:"pools"`
}

type Config struct {
	Dexes []Dex `yaml:"dex"`
	// Tokens  []Token
	Blockchain Blockchain `yaml:"blockchain"`
	General    General    `yaml:"general"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile("config.yaml")
	if err := v.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("unable to read config file")
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		log.Error().Err(err).Msg("unable to unmarshall config file")
	}
	networkres := "blockchain network: " + config.Blockchain.Network
	moderes := "mode: " + config.General.Mode
	log.Info().Msg(networkres)
	log.Warn().Msg(moderes)
	log.Info().Msg("config loaded!")

	return &config, nil

}
