package main

import (
	"dexbot/client"
	"dexbot/config"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Loading configuration")
	config, _ := config.LoadConfig()
	log.Info().Msg("Setting up influx db")
	db := influxdb2.NewClient("http://localhost:8086", "uN9sTAh0apMKfSs6OLPvapkibhU9Uxxg8c8tsahy13kIwH0x7qTahUzhO0raXHq_gWOi0CBPB27ijNkn0OTTog==")
	writeAPI := db.WriteAPIBlocking("PVL", "dexbot")
	log.Info().Msg("Connecting ethereum node")
	client.Connect(config, writeAPI)

}
