/*
   ____             _              _                      _____           _                  _
  / __ \           | |            | |                    |  __ \         | |                | |
 | |  | |_ __   ___| |     ___  __| | __ _  ___ _ __     | |__) | __ ___ | |_ ___   ___ ___ | |
 | |  | | '_ \ / _ \ |    / _ \/ _` |/ _` |/ _ \ '__|    |  ___/ '__/ _ \| __/ _ \ / __/ _ \| |
 | |__| | | | |  __/ |___|  __/ (_| | (_| |  __/ |       | |   | | | (_) | || (_) | (_| (_) | |
  \____/|_| |_|\___|______\___|\__,_|\__, |\___|_|       |_|   |_|  \___/ \__\___/ \___\___/|_|
                                      __/ |
                                     |___/

	Copyright 2017 - 2019 OneLedger

*/

package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Oneledger/protocol/client"
	"github.com/Oneledger/protocol/data/accounts"
	"github.com/Oneledger/protocol/data/chain"
	"github.com/Oneledger/protocol/data/keys"
)

var updateCmd = &cobra.Command{
	Use:   "account",
	Short: "handling an account",
	RunE:  UpdateAccount,
}

// Arguments to the command
type UpdateArguments struct {
	account         string
	chain           string
	pubkey          []byte
	privkey         []byte
	privKeyFilePath string
	delete          []byte
}

var updateArgs = &UpdateArguments{}

func init() {
	RootCmd.AddCommand(updateCmd)

	// Transaction Parameters
	updateCmd.Flags().StringVar(&updateArgs.account, "name", "", "Account Name")
	updateCmd.Flags().StringVar(&updateArgs.chain, "chain", "OneLedger", "Specify the chain")

	updateCmd.Flags().BytesHexVar(&updateArgs.pubkey, "pubkey", []byte{}, "Specify a base64 public key")
	updateCmd.Flags().BytesHexVar(&updateArgs.privkey, "privkey", []byte{}, "Specify a base64 private key")
	updateCmd.Flags().StringVar(&updateArgs.privKeyFilePath, "privKeyFilePath", "", "filepath to save the private key")
	updateCmd.Flags().BytesHexVar(&updateArgs.delete, "delete", []byte{},
		"specify the address of the account to be delete, warning: you can't get back the token it holds")

}

func UpdateAccount(cmd *cobra.Command, args []string) error {

	Ctx := NewContext()
	Ctx.logger.Debug("UPDATING ACCOUNT")

	typ, err := chain.TypeFromName(updateArgs.chain)
	if err != nil {
		Ctx.logger.Error("chain not registered: ", updateArgs.chain)
		//return
	}

	fullnode := Ctx.clCtx.FullNodeClient()

	if len(updateArgs.delete) > 0 {
		_, err := fullnode.DeleteAccount(client.DeleteAccountRequest{Address: updateArgs.delete})
		if err != nil {
			Ctx.logger.Error("delete error", err)
			return err
		}
		Ctx.logger.Info("delete success: ", keys.Address(updateArgs.delete).String())
		return nil
	}
	// get the kys for the new account
	var privKey keys.PrivateKey
	var pubKey keys.PublicKey
	generatedKeysFlag := false
	if len(updateArgs.privkey) == 0 || len(updateArgs.pubkey) == 0 {
		// if a public key or a private key is not passed; generate a pair of keys
		pubKey, privKey, err = keys.NewKeyPairFromTendermint()
		if err != nil {
			Ctx.logger.Error("error generating key from tendermint", err)
		}

		generatedKeysFlag = true
	} else {
		// parse keys passed through commandline

		pubKey, err = keys.GetPublicKeyFromBytes(updateArgs.pubkey, keys.ED25519)
		if err != nil {
			Ctx.logger.Error("incorrect public key", err)
			return err
		}

		privKey, err = keys.GetPrivateKeyFromBytes(updateArgs.privkey, keys.ED25519)
		if err != nil {
			Ctx.logger.Error("incorrect private key", err)
			return err
		}
	}

	// create the account
	acc, err := accounts.NewAccount(typ, updateArgs.account, &privKey, &pubKey)
	if err != nil {
		Ctx.logger.Error("Error initializing account", err)
		return err
	}

	Ctx.logger.Infof("creating account %#v", acc)
	out, err := fullnode.AddAccount(acc)
	if err != nil {
		Ctx.logger.Error("Problem creating account:", err)
	}

	// print details
	Ctx.logger.Infof("Created account successfully: %#v", out)
	Ctx.logger.Infof("Address for the account is: %s", acc.Address().Humanize())

	// if keys are not autogenerated, skip writing private key to file
	if !generatedKeysFlag {
		return nil
	}

	filename := updateArgs.privKeyFilePath
	if filename == "" {
		filename = fmt.Sprintf("./%s_secret", acc.Name)
	}
	writePrivateKeyToFile(Ctx, acc, filename)
	return nil
}

// writePrivateKeyToFile saves a base64 encoded copy of an account secret key to a filepath
func writePrivateKeyToFile(Ctx *Context, acc accounts.Account, filepath string) {

	pkHandler, err := acc.PrivateKey.GetHandler()
	if err != nil {
		Ctx.logger.Error("error getting private key handler", err)
		return
	}

	// open file
	f, err := os.Create(filepath)
	if err != nil {
		Ctx.logger.Error("error opening file for secret", err)
		return
	}

	// pipe base64 encoder to file
	encoder := base64.NewEncoder(base64.StdEncoding, f)
	_, err = encoder.Write(pkHandler.Bytes())
	if err != nil {
		Ctx.logger.Error("error writing bytes to file", err)
		return
	}
	err = encoder.Close()
	if err != nil {
		Ctx.logger.Error("error closing ", err)
	}

	err = f.Close()
	if err != nil {
		Ctx.logger.Error("error closing file ", filepath, "err", err)
	}

	Ctx.logger.Info("Private key wrote to file ", filepath, " successfully.")
}