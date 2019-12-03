package eth

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/Oneledger/protocol/action"
	"github.com/Oneledger/protocol/chains/ethereum"
	"github.com/Oneledger/protocol/data/balance"
	trackerlib "github.com/Oneledger/protocol/data/ethereum"
)

type ReportFinality struct {
	TrackerName      ethereum.TrackerName
	Locker           action.Address
	ValidatorAddress action.Address
	VoteIndex        int64
	Refund           bool
}

var _ action.Msg = &ReportFinality{}

func (m *ReportFinality) Signers() []action.Address {
	return []action.Address{
		m.ValidatorAddress,
	}
}

func (m *ReportFinality) Type() action.Type {
	return action.ETH_REPORT_FINALITY_MINT
}

func (m *ReportFinality) Tags() common.KVPairs {
	tags := make([]common.KVPair, 0)

	tag := common.KVPair{
		Key:   []byte("tx.type"),
		Value: []byte(action.ETH_REPORT_FINALITY_MINT.String()),
	}
	tag2 := common.KVPair{
		Key:   []byte("tx.owner"),
		Value: m.Locker.Bytes(),
	}
	tag3 := common.KVPair{
		Key:   []byte("tx.tracker_name"),
		Value: []byte(m.TrackerName.Hex()),
	}
	tag4 := common.KVPair{
		Key:   []byte("tx.validator"),
		Value: m.ValidatorAddress.Bytes(),
	}

	tags = append(tags, tag, tag2, tag3, tag4)
	return tags
}

func (m *ReportFinality) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *ReportFinality) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

var _ action.Tx = reportFinalityMintTx{}

type reportFinalityMintTx struct {
}

func (r reportFinalityMintTx) Validate(ctx *action.Context, signedTx action.SignedTx) (bool, error) {
	fmt.Println("Starting Validate ")
	f := &ReportFinality{}
	err := f.Unmarshal(signedTx.Data)
	if err != nil {
		return false, errors.Wrap(action.ErrWrongTxType, err.Error())
	}

	err = action.ValidateBasic(signedTx.RawBytes(), f.Signers(), signedTx.Signatures)
	if err != nil {
		return false, err
	}

	if f.VoteIndex < 0 {
		return false, action.ErrMissingData
	}

	return true, nil
}

func (r reportFinalityMintTx) ProcessCheck(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	fmt.Println("START CHECK TX CheckFinality")
	return runCheckFinality(ctx, tx)
}

func (r reportFinalityMintTx) ProcessDeliver(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	fmt.Println("START DELIVER TX CheckFinality")
	return runCheckFinality(ctx, tx)
}

func runCheckFinality(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	fmt.Println("Starting runCheck Finality Mint Internal Trasaction")
	f := &ReportFinality{}
	err := f.Unmarshal(tx.Data)
	if err != nil {
		return false, action.Response{Log: "wrong tx type"}
	}

	tracker, err := ctx.ETHTrackers.Get(f.TrackerName)
	if err != nil {
		ctx.Logger.Error(err, "err getting tracker")
		//	return false, action.Response{Log: err.Error()}
	}

	//
	if tracker.Finalized() {
		return true, action.Response{Log: "tracker already finalized"}
	}

	ctx.Logger.Error("Trying to add vote ")
	index, ok := tracker.CheckIfVoted(f.ValidatorAddress)
	ctx.Logger.Info("Before voting", ok, "Index :", index, "F.index", f.VoteIndex)
	err = tracker.AddVote(f.ValidatorAddress, f.VoteIndex, true)
	if err != nil {
		return false, action.Response{Log: errors.Wrap(err, "failed to add vote").Error()}
	}
	fmt.Printf("%b \n", tracker.FinalityVotes)

	if tracker.Finalized() {
		if tracker.Type == trackerlib.ProcessTypeLock {
			err := mintTokens(ctx, tracker, *f)
			if err != nil {
				return false, action.Response{Log: errors.Wrap(err, "unable to mint tokens").Error()}
			}
		} else if tracker.Type == trackerlib.ProcessTypeRedeem {
			err := burnTokens(ctx, tracker, *f)
			if err != nil {
				return false, action.Response{Log: errors.Wrap(err, "unable to burn tokens").Error()}
			}
		}

		return true, action.Response{Log: "minting successful"}
	}

	err = ctx.ETHTrackers.Set(tracker)
	if err != nil {
		ctx.Logger.Info("Unable to save the tracker", err)
		return false, action.Response{Log: errors.Wrap(err, "unable to save the tracker").Error()}
	}
	// fmt.Println("TRACKER SAVED AT CHECK FINALITY (VOTES): ", tracker.GetVotes())
	ctx.Logger.Info("Voting Done ,unable to mint yet")
	yes, no := tracker.GetVotes()
	return true, action.Response{Log: "vote success, not ready to mint: " + strconv.Itoa(yes) + strconv.Itoa(no)}
}

func (reportFinalityMintTx) ProcessFee(ctx *action.Context, signedTx action.SignedTx, start action.Gas, size action.Gas) (bool, action.Response) {
	return true, action.Response{}
}

func mintTokens(ctx *action.Context, tracker *trackerlib.Tracker, oltTx ReportFinality) error {
	curr, ok := ctx.Currencies.GetCurrencyByName("ETH")
	if !ok {
		return errors.New("ETH currency not allowed")
	}
	lockAmount, err := ethereum.ParseLock(tracker.SignedETHTx)
	if err != nil {
		return err
	}

	oEthCoin := curr.NewCoinFromAmount(*balance.NewAmountFromBigInt(lockAmount.Amount))
	err = ctx.Balances.AddToAddress(oltTx.Locker, oEthCoin)
	if err != nil {
		ctx.Logger.Error(err)
		return errors.New("Unable to mint")
	}

	tracker.State = trackerlib.Released
	err = ctx.ETHTrackers.Set(tracker)
	if err != nil {
		return err
	}
	return nil
}

func burnTokens(ctx *action.Context, tracker *trackerlib.Tracker, oltTx ReportFinality) error {
	curr, ok := ctx.Currencies.GetCurrencyByName("ETH")
	if !ok {
		return errors.New("ETH currency not allowed")
	}
	burnAmount, err := ethereum.ParseRedeem(tracker.SignedETHTx)
	if err != nil {
		return err
	}

	tracker.State = trackerlib.Released
	err = ctx.ETHTrackers.Set(tracker)
	if err != nil {
		return err
	}
	fmt.Println(curr, burnAmount)
	return nil
}
