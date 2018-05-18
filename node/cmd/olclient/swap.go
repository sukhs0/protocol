/*
	Copyright 2017-2018 OneLedger

	Cli to interact with a with the chain.
*/
package main

import (
	"os"

	"github.com/Oneledger/protocol/node/action"
	"github.com/Oneledger/protocol/node/app"
	"github.com/Oneledger/protocol/node/convert"
	"github.com/Oneledger/protocol/node/data"
	"github.com/Oneledger/protocol/node/global"
	"github.com/Oneledger/protocol/node/id"
	"github.com/Oneledger/protocol/node/log"
	"github.com/spf13/cobra"
)

var swapCmd = &cobra.Command{
	Use:   "swap",
	Short: "Setup or confirm a currency swap",
	Run:   SwapCurrency,
}

// Arguments to the command
type SwapArguments struct {
	user       string
	to         string
	from       string
	amount     string
	fee        string
	gas        string // TODO: Not sure this is necessary, unless the chain is like Ethereum
	currency   string
	exchange   string
	excurrency string
	sequence   int // Replay protection
}

var swapargs = &SwapArguments{}

func init() {
	RootCmd.AddCommand(swapCmd)

	// Operational Parameters
	// TODO: Should be global flags?
	swapCmd.Flags().StringVarP(&global.Current.Transport, "transport", "t", "socket", "transport (socket | grpc)")
	swapCmd.Flags().StringVarP(&global.Current.Address, "address", "a", "tcp://127.0.0.1:46658", "full address")

	// Transaction Parameters
	swapCmd.Flags().StringVarP(&swapargs.user, "user", "u", "unknown", "user name")
	swapCmd.Flags().StringVarP(&swapargs.from, "from", "f", "unknown", "base address")
	swapCmd.Flags().StringVarP(&swapargs.to, "to", "d", "unknown", "target address")
	swapCmd.Flags().StringVarP(&swapargs.amount, "amount", "v", "100", "the coins to exchange")
	swapCmd.Flags().StringVarP(&swapargs.fee, "fee", "c", "1", "fees in coins")
	swapCmd.Flags().StringVarP(&swapargs.gas, "gas", "g", "1", "gas, if necessary")
	swapCmd.Flags().StringVarP(&swapargs.currency, "currency", "x", "OLT", "currency of amount")
	swapCmd.Flags().StringVarP(&swapargs.exchange, "exchange", "e", "0", "the value to trade for")
	swapCmd.Flags().StringVarP(&swapargs.excurrency, "excurrency", "y", "ETH", "the currency")
	swapCmd.Flags().IntVarP(&swapargs.sequence, "sequence", "s", 1, "replay seqeunce number")
}

func CreateSwapRequest() []byte {
	log.Debug("swap args", "swapargs", swapargs)

	// TODO: Need better validation and error handling...

	conv := convert.NewConvert()

	party1 := id.Address(conv.GetHash(swapargs.to))
	party2 := id.Address(conv.GetHash(swapargs.from))

	// TOOD: a clash with the basic data model
	signers := GetSigners()

	fee := data.Coin{
		Currency: conv.GetCurrency(swapargs.currency),
		Amount:   conv.GetInt64(swapargs.fee),
	}

	gas := data.Coin{
		Currency: conv.GetCurrency(swapargs.currency),
		Amount:   conv.GetInt64(swapargs.gas),
	}

	amount := data.Coin{
		Currency: conv.GetCurrency(swapargs.currency),
		Amount:   conv.GetInt64(swapargs.amount),
	}

	exchange := data.Coin{
		Currency: conv.GetCurrency(swapargs.excurrency),
		Amount:   conv.GetInt64(swapargs.exchange),
	}

	if conv.HasErrors() {
		Console.Error(conv.GetErrors())
		os.Exit(-1)
	}

	swap := &action.Swap{
		TransactionBase: action.TransactionBase{
			Type:     action.SWAP,
			ChainId:  app.ChainId,
			Signers:  signers,
			Sequence: swapargs.sequence,
		},
		Party1:   party1,
		Party2:   party2,
		Fee:      fee,
		Gas:      gas,
		Amount:   amount,
		Exchange: exchange,
	}

	signed := SignTransaction(action.Transaction(swap))
	packet := PackRequest(signed)

	return packet
}

// IssueRequest sends out a sendTx to all of the nodes in the chain
func SwapCurrency(cmd *cobra.Command, args []string) {
	log.Debug("Swap Request", "tx", swapargs)

	// Create message
	packet := CreateSwapRequest()

	result := Broadcast(packet)

	log.Debug("Returned Successfully", "result", result)
}

func GetAddress(value string) id.Address {
	return id.Address{}
}

func GetCurrency(value string) string {
	// TODO: Check to see that this is a valid currency
	return value
}

func GetInteger(value string) int64 {
	return -1
}
