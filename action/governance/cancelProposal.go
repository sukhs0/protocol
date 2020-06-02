package governance

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/Oneledger/protocol/action"
	gov "github.com/Oneledger/protocol/data/governance"
	"github.com/Oneledger/protocol/data/keys"
)

var _ action.Msg = &CancelProposal{}

type CancelProposal struct {
	ProposalId gov.ProposalID
	Proposer   keys.Address
	Reason     string
}

var _ action.Tx = CancelProposalTx{}

type CancelProposalTx struct {
}

func (c CancelProposalTx) Validate(ctx *action.Context, tx action.SignedTx) (bool, error) {
	ctx.Logger.Debug("Validate CancelProposalTx transaction for CheckTx", tx)

	cc := &CancelProposal{}
	err := cc.Unmarshal(tx.Data)
	if err != nil {
		return false, errors.Wrap(action.ErrWrongTxType, err.Error())
	}

	// validate basic signature
	err = action.ValidateBasic(tx.RawBytes(), cc.Signers(), tx.Signatures)
	if err != nil {
		return false, err
	}

	err = action.ValidateFee(ctx.FeePool.GetOpt(), tx.Fee)
	if err != nil {
		return false, err
	}

	// validate params
	if err = cc.ProposalId.Err(); err != nil {
		return false, errors.Wrap(err, "invalid proposal id")
	}
	if err = cc.Proposer.Err(); err != nil {
		return false, errors.Wrap(err, "invalid proposer address")
	}

	return true, nil
}

func (c CancelProposalTx) ProcessCheck(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	ctx.Logger.Debug("ProcessCheck CancelProposalTx transaction for CheckTx", tx)
	return runCancel(ctx, tx)
}

func (c CancelProposalTx) ProcessFee(ctx *action.Context, signedTx action.SignedTx, start action.Gas, size action.Gas) (bool, action.Response) {
	return action.BasicFeeHandling(ctx, signedTx, start, size, 2)
}

func (c CancelProposalTx) ProcessDeliver(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	ctx.Logger.Debug("ProcessDeliver CancelProposalTx transaction for DeliverTx", tx)
	return runCancel(ctx, tx)
}

func runCancel(ctx *action.Context, tx action.RawTx) (bool, action.Response) {
	cc := &CancelProposal{}
	err := cc.Unmarshal(tx.Data)
	if err != nil {
		return false, action.Response{Log: "calcel proposal failed, deserialization err"}
	}

	// Get Proposal from proposal ACTIVE store
	pms := ctx.ProposalMasterStore
	proposal, err := pms.Proposal.WithPrefixType(gov.ProposalStateActive).Get(cc.ProposalId)
	if err != nil {
		return false, action.Response{
			Log: fmt.Sprintf("calcel proposal failed, id= %v, proposal not found in ACTIVE store", cc.ProposalId)}
	}

	// Check if proposal is in FUNDING status
	if proposal.Status != gov.ProposalStatusFunding {
		return false, action.Response{
			Log: fmt.Sprintf("cancel proposal failed, id= %v, proposal not in FUNDING status", cc.ProposalId)}
	}

	// Check if proposal funding height is passed
	if ctx.Header.Height > proposal.FundingDeadline {
		return false, action.Response{
			Log: fmt.Sprintf("cancel proposal failed, id= %v, funding height passed", cc.ProposalId)}
	}

	// Check if proposer matches
	if !proposal.Proposer.Equal(cc.Proposer) {
		return false, action.Response{
			Log: fmt.Sprintf("cancel proposal failed, id= %v, proposer does not match", cc.ProposalId)}
	}

	// Update fields and add it to FAILED store
	proposal.Status = gov.ProposalStatusCompleted
	proposal.Outcome = gov.ProposalOutcomeCancelled
	proposal.Description += " - Canceled"
	if cc.Reason != "" {
		proposal.Description += fmt.Sprintf("for reason: %v", cc.Reason)
	}
	err = pms.Proposal.WithPrefixType(gov.ProposalStateFailed).Set(proposal)
	if err != nil {
		return false, action.Response{
			Log: fmt.Sprintf("cancel proposal failed, id= %v, failed to add proposal to FAILED store", cc.ProposalId)}
	}

	// Delete proposal in ACTIVE store
	ok, err := pms.Proposal.WithPrefixType(gov.ProposalStateActive).Delete(cc.ProposalId)
	if err != nil || !ok {
		return false, action.Response{
			Log: fmt.Sprintf("cancel proposal failed, id= %v, failed to delete proposal from ACTIVE store", cc.ProposalId)}
	}

	return true, action.Response{Events: action.GetEvent(cc.Tags(), "cancel_proposal_success")}
}

func (cc CancelProposal) Signers() []action.Address {
	return []action.Address{cc.Proposer.Bytes()}
}

func (cc CancelProposal) Type() action.Type {
	return action.PROPOSAL_CANCEL
}

func (cc CancelProposal) Tags() kv.Pairs {
	tags := make([]kv.Pair, 0)

	tag := kv.Pair{
		Key:   []byte("tx.type"),
		Value: []byte(cc.Type().String()),
	}
	tag2 := kv.Pair{
		Key:   []byte("tx.proposalID"),
		Value: []byte(cc.ProposalId),
	}
	tag3 := kv.Pair{
		Key:   []byte("tx.proposer"),
		Value: cc.Proposer.Bytes(),
	}
	tag4 := kv.Pair{
		Key:   []byte("tx.reason"),
		Value: []byte(cc.Reason),
	}

	tags = append(tags, tag, tag2, tag3, tag4)
	return tags
}

func (cc *CancelProposal) Marshal() ([]byte, error) {
	return json.Marshal(cc)
}

func (cc *CancelProposal) Unmarshal(data []byte) error {
	return json.Unmarshal(data, cc)
}
