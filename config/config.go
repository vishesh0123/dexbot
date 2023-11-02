package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Token struct {
	Name     string `yaml:"name"`
	Contract string `yaml:"address"`
	Decimals int    `yaml:"decimals"`
}

type General struct {
	Mode     string `yaml:"mode"`
	LogLevel string `yaml:"logLevel"`
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
	Contract string `yaml:"contract"`
	Type     int    `yaml:"type"`
	Pairs    []Pair `yaml:"pairs"`
}

type AllowedPairs struct {
	Token0 string `yaml:"token0"`
	Token1 string `yaml:"token1"`
}

type Config struct {
	Dexes        []Dex          `yaml:"dexes"`
	Tokens       []Token        `yaml:"tokens"`
	Blockchain   Blockchain     `yaml:"blockchain"`
	General      General        `yaml:"general"`
	AllowedPairs []AllowedPairs `yaml:"allowedpairs"`
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
	networkres := "Blockchain network: " + config.Blockchain.Network
	moderes := "Mode: " + config.General.Mode
	log.Info().Msg(networkres)
	log.Warn().Msg(moderes)
	log.Info().Msg("Config loaded!")
	return &config, nil

}
