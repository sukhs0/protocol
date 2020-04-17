package fees

import (
	"errors"
	"math/big"

	"github.com/Oneledger/protocol/data/balance"
)

const (
	POOL_KEY        = "fee_pool_key"
	FEE_LOCK_BLOCKS = int64(200000)
)

var (
	errNotEnoughMaturedRewards = errors.New("not enough matured rewards")
)

type FeeOption struct {
	FeeCurrency   balance.Currency `json:"feeCurrency"`
	MinFeeDecimal int64            `json:"minFeeDecimal"`

	minimalFee *balance.Coin
}

func (fo *FeeOption) MinFee() balance.Coin {
	if fo.minimalFee == nil {
		amount := balance.Amount(*big.NewInt(0).Exp(big.NewInt(10), big.NewInt(fo.FeeCurrency.Decimal-fo.MinFeeDecimal), nil))
		coin := fo.FeeCurrency.NewCoinFromAmount(amount)
		fo.minimalFee = &coin
	}
	return *fo.minimalFee
}
