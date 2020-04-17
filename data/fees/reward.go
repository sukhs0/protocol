package fees

import (
	"math/big"

	"github.com/Oneledger/protocol/data/balance"
)

type Reward struct {
	Total   *balance.Amount
	Matured *balance.Amount
}

func NewReward(amt *balance.Amount) *Reward {
	return &Reward{
		Total:   amt,
		Matured: balance.NewAmountFromBigInt(big.NewInt(0)),
	}
}

func (r *Reward) Add(amt *balance.Amount) *Reward {
	r.Total = balance.NewAmountFromBigInt(big.NewInt(0).Add(r.Total.BigInt(), amt.BigInt()))
	return r
}

func (r *Reward) Minus(amt *balance.Amount) (*Reward, error) {
	matured := big.NewInt(0).Sub(r.Matured.BigInt(), amt.BigInt())
	if matured.Cmp(big.NewInt(0)) == -1 {
		return r, errNotEnoughMaturedRewards
	}
	//total will always be bigger than matured, so no need to check
	total := big.NewInt(0).Sub(r.Total.BigInt(), amt.BigInt())
	r.Total = balance.NewAmountFromBigInt(total)
	r.Matured = balance.NewAmountFromBigInt(matured)
	return r, nil
}

func (r *Reward) Maturing(amt *balance.Amount) (*Reward, error) {
	r.Matured = balance.NewAmountFromBigInt(big.NewInt(0).Add(r.Matured.BigInt(), amt.BigInt()))
	return r, nil
}
