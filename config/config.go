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
type Pairs struct {
	Token0    string `yaml:"token0"`
	Token1    string `yaml:"token1"`
	Decimals0 uint8  `yaml:"decimals0"`
	Decimals1 uint8  `yaml:"decimals1"`
	Pools     []Pool `yaml:"pools"`
}
type Pool struct {
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Type     uint8  `yaml:"type"`
}

type Config struct {
	Tokens     []Token    `yaml:"tokens"`
	Blockchain Blockchain `yaml:"blockchain"`
	General    General    `yaml:"general"`
	Pairs      []Pairs    `yaml:"pairs"`
}

type EtherscanGasOracleResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		LastBlock       string `json:"LastBlock"`
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
		SuggestBaseFee  string `json:"suggestBaseFee"`
		GasUsedRatio    string `json:"gasUsedRatio"`
		// Add other fields if necessary
	} `json:"result"`
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
