package client

import (
	"errors"

	"github.com/Oneledger/protocol/rpc"
)

// A type-safe client for accessing rpc services.
// Eventually each service will be broken out onto its own type
// TODO: The methods defined here should handle context.Context
type ServiceClient struct {
	*rpc.Client
}

func (c *ServiceClient) Balance(request BalanceRequest) (out BalanceReply, err error) {
	if len(request) <= 20 {
		return out, errors.New("address has insufficient length")
	}
	err = c.Call("query.Balance", request, &out)
	return
}

func (c *ServiceClient) NodeName() (out NodeNameReply, err error) {
	err = c.Call("node.Name", struct{}{}, &out)
	return
}

func (c *ServiceClient) NodeAddress() (out NodeAddressReply, err error) {
	err = c.Call("node.Address", struct{}{}, &out)
	return
}

func (c *ServiceClient) NodeID(req NodeIDRequest) (out NodeIDReply, err error) {
	err = c.Call("node.ID", req, &out)
	return
}

func (c *ServiceClient) SendTx(req SendTxRequest) (out SendTxReply, err error) {
	err = c.Call("tx.SendTx", req, &out)
	return
}

func (c *ServiceClient) AddAccount(req AddAccountRequest) (out AddAccountReply, err error) {
	err = c.Call("owner.AddAccount", req, &out)
	return
}

func (c *ServiceClient) DeleteAccount(req DeleteAccountRequest) (out DeleteAccountReply, err error) {
	err = c.Call("owner.DeleteAccount", req, &out)
	return
}

func (c *ServiceClient) ListAccounts() (out ListAccountsReply, err error) {
	err = c.Call("owner.ListAccounts", struct{}{}, &out)
	return
}

func (c *ServiceClient) ApplyValidator(req ApplyValidatorRequest) (out ApplyValidatorReply, err error) {
	err = c.Call("tx.ApplyValidator", req, &out)
	return
}

/* ONS */
func (c *ServiceClient) ONS_CreateRawCreate(req ONSCreateRequest) (out SendTxReply, err error) {
	err = c.Call("tx.ONS_CreateRawCreate", req, &out)
	return
}
func (c *ServiceClient) ONS_CreateRawUpdate(req ONSUpdateRequest) (out SendTxReply, err error) {
	err = c.Call("tx.ONS_CreateRawUpdate", req, &out)
	return
}
func (c *ServiceClient) ONS_CreateRawSale(req ONSSaleRequest) (out SendTxReply, err error) {
	err = c.Call("tx.ONS_CreateRawSale", req, &out)
	return
}
func (c *ServiceClient) ONS_CreateRawBuy(req ONSPurchaseRequest) (out SendTxReply, err error) {
	err = c.Call("tx.ONS_CreateRawBuy", req, &out)
	return
}
func (c *ServiceClient) ONS_CreateRawSend(req ONSSendRequest) (out SendTxReply, err error) {
	err = c.Call("tx.ONS_CreateRawSend", req, &out)
	return
}

func (c *ServiceClient) ListValidators() (out ListValidatorsReply, err error) {
	err = c.Call("query.ListValidators", struct{}{}, &out)
	return
}

/* Broadcast */
func (c *ServiceClient) TxAsync(req BroadcastRequest) (out BroadcastReply, err error) {
	err = c.Call("broadcast.TxAsync", req, &out)
	return

}

func (c *ServiceClient) TxSync(req BroadcastRequest) (out BroadcastReply, err error) {
	err = c.Call("broadcast.TxSync", req, &out)
	return
}

func (c *ServiceClient) TxCommit(req BroadcastRequest) (out BroadcastReply, err error) {
	err = c.Call("broadcast.TxCommit", req, &out)
	return
}
