package client

import (
	"context"
	"dexbot/abi"
	"dexbot/config"
	"dexbot/store"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
)

func Connect(conf *config.Config, api api.WriteAPIBlocking) error {
	client, err := rpc.Dial("wss://eth-mainnet.g.alchemy.com/v2/zyHryxjXbmPUbfmKbvzj8YDw_5zharB1")

	if err != nil {
		log.Error().Err(err).Msg("unable to connect ethereum client")
	}
	headers := make(chan *types.Header)
	sub, err := client.EthSubscribe(context.Background(), headers, "newHeads")

	if err != nil {
		log.Error().Err(err).Msg("failed to subscribe")
	}

	for {
		select {
		case err := <-sub.Err():
			log.Error().Err(err).Msg("subscription error")
		case h := <-headers:
			// A new block header has arrived, run the necessary function
			log.Info().Msg("new block mined")
			fmt.Println(h.Hash().Hex(), "timestamp ", h.Time)

			allowed := conf.AllowedPairs
			for _, v := range allowed {
				for _, dex := range store.GlobalDexStore {
					pool := store.GlobalPoolStore[store.PairWithoutPool(v)][dex]
					pair := config.Pair{
						Token0: v.Token0,
						Token1: v.Token1,
						Pool:   pool,
					}
					var price float64
					if dex == "UNISWAPV2" || dex == "SUSHISWAPV2" {
						price, _ = FetchPriceV2(pair)
					} else {
						price, _ = FetchPriceV3(pair)
					}
					p := influxdb2.NewPointWithMeasurement("new").
						AddTag("exchange", dex).
						AddTag("pair", v.Token0+"/"+v.Token1).
						AddField("price", 1/price)
					go api.WritePoint(context.Background(), p)

				}

			}
		}
	}

	return nil
}

func FetchPriceV3(pair config.Pair) (float64, error) {
	conn, err := ethclient.Dial("wss://eth-mainnet.g.alchemy.com/v2/zyHryxjXbmPUbfmKbvzj8YDw_5zharB1")
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect ethereum client")
	}
	pool, _ := abi.NewPool(common.HexToAddress(pair.Pool), conn)
	slot0, _ := pool.Slot0(nil)
	tick, _ := slot0.Tick.Float64()
	difference := store.GlobalDecimals[pair.Token1] - store.GlobalDecimals[pair.Token0]
	price := math.Pow(1.0001, tick) / math.Pow10(difference)
	return price, nil

}

func FetchPriceV2(pair config.Pair) (float64, error) {
	conn, err := ethclient.Dial("wss://eth-mainnet.g.alchemy.com/v2/zyHryxjXbmPUbfmKbvzj8YDw_5zharB1")
	if err != nil {
		log.Error().Err(err).Msg("Unable to connect ethereum client")
	}
	pool, _ := abi.NewPOOLV2(common.HexToAddress(pair.Pool), conn)
	reserves, _ := pool.GetReserves(nil)
	token0, _ := reserves.Reserve0.Float64()
	token1, _ := reserves.Reserve1.Float64()
	t0 := token0 / math.Pow10(store.GlobalDecimals[pair.Token0])
	t1 := token1 / math.Pow10(store.GlobalDecimals[pair.Token1])
	price := t1 / t0
	return price, nil

}
