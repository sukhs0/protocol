package governance

import (
	"encoding/json"
	"github.com/Oneledger/protocol/action"
	"github.com/Oneledger/protocol/data/balance"
	"github.com/Oneledger/protocol/data/governance"
	"github.com/Oneledger/protocol/data/keys"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/kv"
	"math/big"
)

var _ action.Msg = &WithdrawProposal{}

type WithdrawProposal struct {
	ProposalID    governance.ProposalID
	Contributor   keys.Address
	WithdrawValue action.Amount
	Beneficiary   keys.Address

}

func (wp WithdrawProposal) Validate(ctx *action.Context, signedTx action.SignedTx) (bool, error) {
	withdrawProposal := WithdrawProposal{}
	err := withdrawProposal.Unmarshal(signedTx.Data)
	if err != nil {
		return false, errors.Wrap(action.ErrWrongTxType, err.Error())
	}

	//validate basic signature
	err = action.ValidateBasic(signedTx.RawBytes(), withdrawProposal.Signers(), signedTx.Signatures)
	if err != nil {
		return false, err
	}
	//validate fee
	err = action.ValidateFee(ctx.FeePool.GetOpt(), signedTx.Fee)
	if err != nil {
		return false, err
	}

	// the currency should be OLT
	currency, ok := ctx.Currencies.GetCurrencyById(0)
	if !ok {
		panic("no default currency available in the network")
	}
	if currency.Name != withdrawProposal.WithdrawValue.Currency {
		return false, errors.Wrap(action.ErrInvalidAmount, withdrawProposal.WithdrawValue.String())
	}

	//Check if fund contributor address is valid oneLedger address
	err = withdrawProposal.Contributor.Err()
	if err != nil {
		return false, errors.Wrap(err, "invalid withdraw contributor address")
	}

	//Check if withdraw beneficiary address is valid oneLedger address
	err = withdrawProposal.Beneficiary.Err()
	if err != nil {
		return false, errors.Wrap(err, "invalid withdraw beneficiary address")
	}

	return true, nil
}

func (wp WithdrawProposal) ProcessCheck(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	ctx.Logger.Debug("Processing WithdrawProposal Transaction for CheckTx", tx)
	return runWithdraw(ctx, tx)
}

func (wp WithdrawProposal) ProcessDeliver(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	ctx.Logger.Debug("Processing WithdrawProposal Transaction for DeliverTx", tx)
	return runWithdraw(ctx, tx)
}

func (wp WithdrawProposal) ProcessFee(ctx *action.Context, signedTx action.SignedTx, start action.Gas, size action.Gas) (bool, action.Response) {
	return action.BasicFeeHandling(ctx, signedTx, start, size, 1)
}

func runWithdraw(ctx *action.Context, signedTx action.RawTx) (bool, action.Response) {
	withdrawProposal := WithdrawProposal{}
	err := withdrawProposal.Unmarshal(signedTx.Data)
	if err != nil {
		return false, action.Response{}
	}

	// 1. Check if Proposal already exists, if so, check the withdraw requirement:
	//    a. the funding goal is not reached
	//    b. the funding height is reached

	proposal, _, err := ctx.ProposalMasterStore.Proposal.QueryAllStores(withdrawProposal.ProposalID)
	if err != nil {
		ctx.Logger.Error("Proposal does not exist :", withdrawProposal.ProposalID)
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "withdraw_proposal_does_not_exist"),
		}
		return false, result
	}
	currentFundsForProposal := governance.GetCurrentFunds(proposal.ProposalID, ctx.ProposalMasterStore.ProposalFund)
	// if funding goal is reached or there is still time for funding
	if currentFundsForProposal.BigInt().Cmp(proposal.FundingGoal.BigInt()) >= 0 || ctx.Header.Height <= proposal.FundingDeadline {
		ctx.Logger.Error("Proposal does not meet withdraw requirement", withdrawProposal.ProposalID)
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "withdraw_proposal_does_not_meet_withdraw_requirement"),
		}
		return false, result
	}
	// 2. change status and outcome
	proposal.Status = governance.ProposalStatusWithdrawing
	proposal.Outcome = governance.ProposalOutcomeInsufficientFunds


	// 3. Check if the contributor has funded this proposal, if so, get the amount of funds
	proposalFund, err := governance.GetCurrentFundsByContributor(proposal.ProposalID, withdrawProposal.Contributor, ctx.ProposalMasterStore.ProposalFund)
	if err != nil {
		ctx.Logger.Error("No available funds to withdraw for this contributor :", withdrawProposal.Contributor)
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "no_available__fund_to_withdraw_for_this_contributor"),
		}
		return false, result
	}

	// 4. Check if the contributor has sufficient funds to withdraw for that proposal
	withdrawAmount := balance.NewAmountFromBigInt(withdrawProposal.WithdrawValue.Value.BigInt())
	if proposalFund.BigInt().Cmp(withdrawAmount.BigInt()) < 0 {
		ctx.Logger.Error("Insufficient funds to withdraw for this contributor :", withdrawProposal.Contributor)
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "no_available__fund_to_withdraw_for_this_contributor"),
		}
		return false, result
	}

	// 5. withdraw
	// deduct from proposal fund
	err = ctx.ProposalMasterStore.ProposalFund.DeductFunds(proposal.ProposalID, withdrawProposal.Contributor, withdrawAmount)
	if err != nil {
		ctx.Logger.Error("Failed to deduct funds from proposal:", withdrawProposal.ProposalID)
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "withdraw_proposal_DeductFund_failed"),
		}
		return false, result
	}
	// add to beneficiary address
	coin := withdrawProposal.WithdrawValue.ToCoin(ctx.Currencies)
	err = ctx.Balances.AddToAddress(withdrawProposal.Beneficiary.Bytes(), coin)
	if err != nil {
		result := action.Response{
			Events: action.GetEvent(withdrawProposal.Tags(), "withdraw_proposal_addition_failed"),
		}
		return false, result
	}

	result := action.Response{
		Events: action.GetEvent(withdrawProposal.Tags(), "withdraw_proposal_success"),
	}

	// check proposal funds:
	// if all funds for this proposal has been withdrawn, mark withdraw complete
	if currentFundsForProposal.BigInt().Cmp(big.NewInt(0)) == 0 {
		proposal.Status = governance.ProposalStatusWithdrawCompleted
		proposal.Outcome = governance.ProposalOutcomeCancelled
	}

	return true, result
}

func (wp WithdrawProposal) Signers() []action.Address {
	return []action.Address{wp.Contributor}
}

func (wp WithdrawProposal) Type() action.Type {
	return action.PROPOSAL_WITHDRAW
}

func (wp WithdrawProposal) Tags() kv.Pairs {
	tags := make([]kv.Pair, 0)

	tag := kv.Pair{
		Key:   []byte("tx.type"),
		Value: []byte(wp.Type().String()),
	}
	tag2 := kv.Pair{
		Key:   []byte("tx.proposalID"),
		Value: []byte(string(wp.ProposalID)),
	}
	tag3 := kv.Pair{
		Key:   []byte("tx.contributor"),
		Value: wp.Contributor.Bytes(),
	}
	tag4 := kv.Pair{
		Key:   []byte("tx.withdrawValue"),
		Value: []byte(wp.WithdrawValue.String()),
	}
	tag5 := kv.Pair{
		Key:   []byte("tx.beneficiary"),
		Value: wp.Beneficiary.Bytes(),
	}


	tags = append(tags, tag, tag2, tag3, tag4, tag5)
	return tags
}

func (wp WithdrawProposal) Marshal() ([]byte, error) {
	return json.Marshal(wp)
}

func (wp WithdrawProposal) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, wp)
}
