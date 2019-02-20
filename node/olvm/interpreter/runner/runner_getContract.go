/*
	Copyright 2017-2018 OneLedger
*/
package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/Oneledger/protocol/node/action"
	"github.com/Oneledger/protocol/node/log"
)

func (runner Runner) setupContract(request *action.OLVMRequest) bool {
	address := request.Address
	sourceCode := ""

	switch {
	case strings.HasPrefix(address, "samples://"):
		sourceCode = getSourceCodeFromSamples(address)
	case address == "embed://":
		// TODO: Should preserve byte array, to support UTF8?
		sourceCode = string(request.SourceCode)
	default:
		sourceCode = getSourceCodeFromBlockChain(address)
	}

	// TODO: Needs better error handling
	if sourceCode == "" {
		return false
	}
	log.Debug("get source code", "sourceCode", sourceCode)
	_, error := runner.vm.Run(`var module = {};(function(module){` + sourceCode + `})(module)`)
	if error == nil {
		return true
	} else {
		return false
	}
}

func getSourceCodeFromSamples(address string) string {

	prefix := "samples://"
	sampleCodeName := address[len(prefix):]
	jsFilePath := filepath.Join(os.Getenv("OLROOT"), "/protocol/node/olvm/interpreter/samples/", sampleCodeName+".js")
	log.Debug("get source code from local file system", "path", jsFilePath)
	file, err := os.Open(jsFilePath)
	if err != nil {

		// TODO: Needs better error handling
		log.Debug("cannot get source code", "err", err)
		return ""
		//log.Fatal(err)
	}

	defer file.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	contents := buf.String()

	return contents
}

func getSourceCodeFromBlockChain(address string) string {
	log.Fatal("SourceCodeFrom BlockChain is Unimplemented")
	return ""
}
