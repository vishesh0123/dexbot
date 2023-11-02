package store

import (
	"dexbot/config"

	"github.com/rs/zerolog/log"
)

type PairWithoutPool struct {
	Token0 string
	Token1 string
}
type PoolStore map[PairWithoutPool]map[string]string
type DexStore []string
type Decimals map[string]int

var GlobalPoolStore PoolStore
var GlobalDexStore DexStore
var GlobalDecimals Decimals

func SetupDexes(config *config.Config) error {
	// storing all pairs and pool
	dexes := config.Dexes

	if GlobalPoolStore == nil {
		GlobalPoolStore = make(PoolStore)
		GlobalDecimals = make(Decimals)
	}

	for _, v := range dexes {
		GlobalDexStore = append(GlobalDexStore, v.Name)
		log.Info().Msg("Loaded " + v.Name)
		// load all pools
		for _, pair := range v.Pairs {
			newPair := PairWithoutPool{
				Token0: pair.Token0,
				Token1: pair.Token1,
			}
			if GlobalPoolStore[newPair] == nil {
				GlobalPoolStore[newPair] = make(map[string]string)
			}

			GlobalPoolStore[newPair][v.Name] = pair.Pool
			log.Info().Msg("Loaded pool " + pair.Token0 + "/" + pair.Token1 + " " + v.Name)
		}
	}

	tokens := config.Tokens

	for _, v := range tokens {
		GlobalDecimals[v.Name] = v.Decimals
	}

	return nil

}
