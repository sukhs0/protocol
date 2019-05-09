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

	"github.com/Oneledger/protocol/data"
	"github.com/Oneledger/protocol/data/accounts"
	"github.com/Oneledger/protocol/data/chain"
	"github.com/Oneledger/protocol/data/keys"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an account",
	Run:   UpdateAccount,
}

// Arguments to the command
type UpdateArguments struct {
	account         string
	chain           string
	pubkey          []byte
	privkey         []byte
	nodeaccount     bool
	privKeyFilePath string
}

var updateArgs = &UpdateArguments{}

func init() {
	RootCmd.AddCommand(updateCmd)

	// Transaction Parameters
	updateCmd.Flags().StringVar(&updateArgs.account, "account", "", "Account Name")
	updateCmd.Flags().StringVar(&updateArgs.chain, "chain", "OneLedger", "Specify the chain")

	updateCmd.Flags().BytesBase64Var(&updateArgs.pubkey, "pubkey", []byte{}, "Specify a base64 public key")
	updateCmd.Flags().BytesBase64Var(&updateArgs.privkey, "privkey", []byte{}, "Specify a base64 private key")
	updateCmd.Flags().BoolVar(&updateArgs.nodeaccount, "nodeaccount", false, "Specify whether it's a node account or not")

	updateCmd.Flags().StringVar(&updateArgs.privKeyFilePath, "privKeyFilePath", "", "filepath to save the private key")
}

func UpdateAccount(cmd *cobra.Command, args []string) {
	logger.Debug("UPDATING ACCOUNT")

	typ, err := chain.TypeFromName(updateArgs.chain)
	if err != nil {
		logger.Error("chain not registered")
		return
	}

	var privKey keys.PrivateKey
	var pubKey keys.PublicKey

	generatedKeysFlag := false
	if len(updateArgs.privkey) == 0 || len(updateArgs.pubkey) == 0 {
		// if a public key or a private key is not passed; generate a pair of keys
		tmPrivKey := ed25519.GenPrivKey()
		tmPublicKey := tmPrivKey.PubKey()

		pubKey, err = keys.GetPublicKeyFromBytes(tmPublicKey.Bytes(), keys.ED25519)
		if err != nil {
			logger.Error("error in generated public key", err)
			return
		}
		privKey, err = keys.GetPrivateKeyFromBytes(tmPrivKey.Bytes(), keys.ED25519)
		if err != nil {
			logger.Error("error in generated private key", err)
		}

		generatedKeysFlag = true
	} else {
		// parse keys passed through commandline

		pubKey, err = keys.GetPublicKeyFromBytes(updateArgs.pubkey, keys.ED25519)
		if err != nil {
			logger.Error("incorrect public key", err)
			return
		}

		privKey, err = keys.GetPrivateKeyFromBytes(updateArgs.privkey, keys.ED25519)
		if err != nil {
			logger.Error("incorrect private key", err)
			return
		}
	}

	acc, err := accounts.NewAccount(typ, updateArgs.account, privKey, pubKey)
	if err != nil {
		logger.Error("Error initializing account", err)
		return
	}

	resp := &data.Response{}
	err = Ctx.Query("AddAcount", acc, resp)
	if err != nil {
		logger.Error("error creating account", err)
		return
	}

	logger.Info("Created account successfully", "account", acc)
	logger.Info("Address for the account is: ", acc.Address())

	// if keys are not autogenerated, skip writing private key to file
	if generatedKeysFlag == false {
		return
	}

	filename := updateArgs.privKeyFilePath
	if filename == "" {
		filename = fmt.Sprintf("./%s_secret", acc.Name)
	}
	writePrivateKeyToFile(acc, filename)

}

// writePrivateKeyToFile saves a base64 encoded copy of an account secret key to a filepath
func writePrivateKeyToFile(acc accounts.Account, filepath string) {

	pkHandler, err := acc.PrivateKey.GetHandler()
	if err != nil {
		logger.Error("error getting private key handler", err)
		return
	}

	// open file
	f, err := os.Create(filepath)
	if err != nil {
		logger.Error("error opening file for secret", err)
		return
	}

	// pipe base64 encoder to file
	encoder := base64.NewEncoder(base64.StdEncoding, f)
	_, err = encoder.Write(pkHandler.Bytes())
	if err != nil {
		logger.Error("error writing bytes to file", err)
		return
	}
	err = encoder.Close()
	if err != nil {
		logger.Error("error closing ", err)
	}

	err = f.Close()
	if err != nil {
		logger.Error("error closing file ", filepath, "err", err)
	}

	logger.Info("Private key wrote to file ", filepath, " successfully.")
}
