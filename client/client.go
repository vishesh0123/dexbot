package client

import (
	"context"
	"dexbot/abi"
	"dexbot/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
)

func Connect(conf *config.Config, api api.WriteAPIBlocking) error {
	calldata := GenerateCalldata(conf)
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
			go fetchPrice(conf, api, calldata)
			go FetchGasPrice(api)

		}
	}

	return nil
}

func FetchPriceV3(tick float64, difference int, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {

	price := math.Pow(1.0001, tick) / math.Pow10(difference)
	// p := influxdb2.NewPointWithMeasurement("new").
	// 	AddTag("exchange", pool.Protocol).
	// 	AddTag("pair", pair.Token0+"/"+pair.Token1).
	// 	AddField("price", 1/price)
	// go api.WritePoint(context.Background(), p)
	log.Info().Msg("V3 L")
	return price, nil

}

func FetchPriceV2(token0 float64, token1 float64, d0 int, d1 int, pair config.Pairs, pool config.Pool, api api.WriteAPIBlocking) (float64, error) {
	t0 := token0 / math.Pow10(d0)
	t1 := token1 / math.Pow10(d1)
	price := t1 / t0
	// p := influxdb2.NewPointWithMeasurement("new").
	// 	AddTag("exchange", pool.Protocol).
	// 	AddTag("pair", pair.Token0+"/"+pair.Token1).
	// 	AddField("price", 1/price)
	// go api.WritePoint(context.Background(), p)
	log.Info().Msg("V2 L")
	return price, nil

}

func fetchPrice(conf *config.Config, api api.WriteAPIBlocking, calldata []abi.Multicall3Call) error {
	conn, err := ethclient.Dial("wss://eth-mainnet.g.alchemy.com/v2/i6wI8GHx5UD_uyN8jjBXxBDs5UfbGLGM")

	if err != nil {
		log.Error().Err(err).Msg("Unable to connect ethereum client")
	}

	multicall, _ := abi.NewMullticall(common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"), conn)

	res, err := multicall.TryAggregate(nil, false, calldata)

	if err != nil {
		log.Error().Err(err).Msg("Failed calling via calldata")
	}

	counter := 0
	for _, v := range conf.Pairs {

		for _, p := range v.Pools {
			data := res[counter].ReturnData
			if p.Type == 1 {
				data := data[20:33]
				tick := float64((int32(int8(data[0]))<<16 | int32(data[1])<<8 | int32(data[2])))
				go FetchPriceV3(tick, int(math.Abs(float64(v.Decimals0)-float64(v.Decimals1))), v, p, api)

			} else {
				t0, _ := new(big.Int).SetBytes(data[0:32]).Float64()
				t1, _ := new(big.Int).SetBytes(data[32:64]).Float64()
				go FetchPriceV2(t0, t1, int(v.Decimals0), int(v.Decimals1), v, p, api)
			}

			counter++

		}
	}
	return nil
}

func GenerateCalldata(conf *config.Config) []abi.Multicall3Call {
	// Generates the calldata required for multicall
	var data []abi.Multicall3Call

	signature1 := []byte("slot0()")
	selector1 := crypto.Keccak256(signature1)[:4]

	signature2 := []byte("getReserves()")
	selector2 := crypto.Keccak256(signature2)[:4]

	for _, v := range conf.Pairs {
		for _, pool := range v.Pools {
			var selector []byte
			if pool.Type == 1 {
				selector = selector1
			} else {
				selector = selector2
			}
			data = append(data, abi.Multicall3Call{
				Target:   common.HexToAddress(pool.Address),
				CallData: selector,
			})
		}

	}
	return data

}

func FetchGasPrice(api api.WriteAPIBlocking) error {
	res, err := http.Get("https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=Q3CDHFRFTH5SVSAD6XQQ841JXK65WC7926")
	if err != nil {
		log.Error().Err(err).Msg("Failed To call gas tracker api")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed To Read Body")
	}
	var gasOracleResponse config.EtherscanGasOracleResponse
	if err := json.Unmarshal(body, &gasOracleResponse); err != nil {
		log.Error().Err(err).Msg("Failed To Unmarshall Body")
	}
	p := influxdb2.NewPointWithMeasurement("gas_price").
		AddTag("Network", "Ethereum").
		AddField("baseGasFees", gasOracleResponse.Result.SuggestBaseFee).
		AddField("low_Price", gasOracleResponse.Result.SafeGasPrice).
		AddField("Avg_Price", gasOracleResponse.Result.ProposeGasPrice).
		AddField("High_Price", gasOracleResponse.Result.FastGasPrice)
	go api.WritePoint(context.Background(), p)
	fmt.Println("Base Gas Price := ", gasOracleResponse.Result.SuggestBaseFee)
	fmt.Println("Low Price: ", gasOracleResponse.Result.SafeGasPrice)
	fmt.Println("Avg_price", gasOracleResponse.Result.ProposeGasPrice)
	fmt.Println("High Price", gasOracleResponse.Result.FastGasPrice)
	fmt.Println("Network Congestion (Gas Ratios)", gasOracleResponse.Result.GasUsedRatio)

	return nil
}
