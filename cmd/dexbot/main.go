package main

import (
	"os"

	"dexbot/client"
	"dexbot/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("loading configuration")
	config.LoadConfig()
	log.Info().Msg("setting up influx db")
	// client := influxdb2.NewClient("http://localhost:8086", "LxAAIlRb-LlgvzCBq-DYBcisb9ml7fpl_oXtShIBilokchdRE5fYpNd40e9St7W8isAy-NwEjc2Fiw70ewCsIA==")
	// writeAPI := client.WriteAPIBlocking("PVL", "dexbot")
	log.Info().Msg("connecting ethereum node")
	client.Connect()

}
