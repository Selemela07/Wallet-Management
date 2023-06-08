package ethereum

import (
	"math"
	"math/big"
)

func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float)
	ether.SetString(wei.String())
	return new(big.Float).Quo(ether, big.NewFloat(math.Pow10(18)))
}
