/*
	Copyright 2017-2018 OneLedger

	Cli to interact with a with the chain.
*/
package main

import (
	"github.com/Oneledger/protocol/node/cmd/shared"
	"github.com/Oneledger/protocol/node/comm"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create or reuse an account",
	Run:   RegisterIdentity,
}

// Arguments to the command
type RegistrationArguments struct {
	identity string
	pubkey   string
}

var arguments = &RegistrationArguments{}

func init() {
	RootCmd.AddCommand(registerCmd)

	// Transaction Parameters
	registerCmd.Flags().StringVar(&arguments.identity, "identity", "Unknown", "User's Identity")
	registerCmd.Flags().StringVar(&arguments.pubkey, "pubkey", "0x00000000", "Specify a public key")
}

func RegisterIdentity(cmd *cobra.Command, args []string) {
	arguments := &shared.RegisterArguments{}

	register := shared.CreateRegisterRequest(arguments)

	comm.SDKRequest(register)
}

/*
// IssueRequest sends out a sendTx to all of the nodes in the chain
func Register(cmd *cobra.Command, args []string) {
	log.Debug("Client Register Account via SetOption...")

	cli := &app.RegisterArguments{
		Identity:   arguments.identity,
		Chain:      arguments.chain,
		PublicKey:  arguments.pubkey,
		PrivateKey: arguments.privkey,
	}

	buffer, err := serial.Serialize(cli, serial.CLIENT)
	if err != nil {
		log.Error("Register Failed", "err", err)
		return
	}
	comm.SetOption("Register", string(buffer))
}
*/
