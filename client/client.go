package client

import (
	"context"
	"dexbot/abi"
	"dexbot/config"
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
			go fetchPrice(conf, api)

		}
	}

	return nil
}

func FetchPriceV3(tick float64, difference int, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {

	price := math.Pow(1.0001, tick) / math.Pow10(difference)
	p := influxdb2.NewPointWithMeasurement("new").
		AddTag("exchange", pool.Protocol).
		AddTag("pair", pair.Token0+"/"+pair.Token1).
		AddField("price", 1/price)
	go api.WritePoint(context.Background(), p)
	return price, nil

}

func FetchPriceV2(token0 float64, token1 float64, d0 int, d1 int, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {
	t0 := token0 / math.Pow10(d0)
	t1 := token1 / math.Pow10(d1)
	price := t1 / t0
	p := influxdb2.NewPointWithMeasurement("new").
		AddTag("exchange", pool.Protocol).
		AddTag("pair", pair.Token0+"/"+pair.Token1).
		AddField("price", 1/price)
	go api.WritePoint(context.Background(), p)
	return price, nil

}

func FetchLiquidityV2(token0 float64, token1 float64, d0 int, d1 int, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {
	t0 := token0 / math.Pow10(d0)
	t1 := token1 / math.Pow10(d1)
	liquidity := math.Sqrt(t0 * t1)
	p := influxdb2.NewPointWithMeasurement("new").
		AddTag("exchange", pool.Protocol).
		AddTag("pair", pair.Token0+"/"+pair.Token1).
		AddField("liquidity", liquidity)
	go api.WritePoint(context.Background(), p)

	return liquidity, nil
}

func FetchLiquidityV3(liquidity float64, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {
	square := liquidity / math.Pow10(int(math.Abs(float64(pair.Decimals0)-float64(pair.Decimals1))))
	p := influxdb2.NewPointWithMeasurement("new").
		AddTag("exchange", pool.Protocol).
		AddTag("pair", pair.Token0+"/"+pair.Token1).
		AddField("liquidity", square)
	go api.WritePoint(context.Background(), p)
	return liquidity, nil
}

func fetchPrice(conf *config.Config, api api.WriteAPIBlocking) error {
	conn, err := ethclient.Dial("wss://eth-mainnet.g.alchemy.com/v2/i6wI8GHx5UD_uyN8jjBXxBDs5UfbGLGM")

	if err != nil {
		log.Error().Err(err).Msg("Unable to connect ethereum client")
	}

	for _, v := range conf.Pairs {
		for _, p := range v.Pools {
			if p.Type == 1 {
				pool, _ := abi.NewPool(common.HexToAddress(p.Address), conn)
				slot, _ := pool.Slot0(nil)
				tick, _ := slot.Tick.Float64()
				liquidity, _ := pool.Liquidity(nil)
				liq, _ := liquidity.Float64()
				go FetchPriceV3(tick, int(math.Abs(float64(v.Decimals1)-float64(v.Decimals0))), v, p, api)
				go FetchLiquidityV3(liq, v, p, api)

			} else {
				pool, _ := abi.NewPOOLV2(common.HexToAddress(p.Address), conn)
				reserves, _ := pool.GetReserves(nil)
				t0, _ := reserves.Reserve0.Float64()
				t1, _ := reserves.Reserve1.Float64()
				go FetchPriceV2(t0, t1, int(v.Decimals0), int(v.Decimals1), v, p, api)
				go FetchLiquidityV2(t0, t1, int(v.Decimals0), int(v.Decimals1), v, p, api)
			}

		}
	}
	return nil
}

// func calculateTypeOne(prices []*big.Int, decimals []uint8) error {
// 	for i, v := range prices {
// 		c, _ := v.Float64()
// 		FetchPriceV3(c, int(decimals[i]))
// 	}
// 	return nil

// }

// func calculateTypeTwo(x []*big.Int, y []*big.Int, dx []uint8, dy []uint8) error {
// 	for i, v := range x {
// 		t0, _ := v.Float64()
// 		t1, _ := y[i].Float64()
// 		FetchPriceV2(t0, t1, int(dx[i]), int(dy[i]))
// 	}
// 	return nil
// }

// func Decimals(conf *config.Config) ([]uint8, []uint8, []uint8) {
// 	var x []uint8
// 	var y []uint8
// 	var d []uint8
// 	for _, v := range conf.Pairs {
// 		for _, pool := range v.Pools {
// 			if pool.Type == 1 {
// 				d = append(d, uint8(math.Abs(float64(v.Decimals0)-float64(v.Decimals1))))
// 			} else {
// 				x = append(x, v.Decimals0)
// 				y = append(y, v.Decimals1)
// 			}
// 		}
// 	}
// 	return x, y, d

// }

// func ToCommonAddress(conf *config.Config) ([]common.Address, []common.Address) {
// 	var type1 []common.Address
// 	var type2 []common.Address
// 	for _, v := range conf.Pairs {
// 		for _, pool := range v.Pools {
// 			if pool.Type == 1 {
// 				type1 = append(type1, common.HexToAddress(pool.Address))
// 			} else {
// 				type2 = append(type2, common.HexToAddress(pool.Address))
// 			}

// 		}

// 	}
// 	return type1, type2

// }
