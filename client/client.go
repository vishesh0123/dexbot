package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog/log"
)

func Connect() error {
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
			fmt.Println(h.Hash().Hex())
		}
	}

	return nil
}
