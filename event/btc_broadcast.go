/*

 */

package event

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	"github.com/Oneledger/protocol/action"
	"github.com/Oneledger/protocol/action/btc"
	"github.com/Oneledger/protocol/chains/bitcoin"
	"github.com/Oneledger/protocol/data/jobs"
)

type JobBTCBroadcast struct {
	Type string

	TrackerName string

	JobID string

	Status jobs.Status
}

func NewBTCBroadcastJob(trackerName string) jobs.Job {

	id := strconv.FormatInt(time.Now().UnixNano(), 10)

	return &JobBTCBroadcast{
		Type:        JobTypeBTCBroadcast,
		TrackerName: trackerName,
		JobID:       id,
		Status:      jobs.New,
	}
}

func (j *JobBTCBroadcast) DoMyJob(ctxI interface{}) {

	ctx, _ := ctxI.(*JobsContext)

	tracker, err := ctx.Trackers.Get(j.TrackerName)
	if err != nil {
		ctx.Logger.Error("err trying to deserialize tracker: ", j.TrackerName, err)
		return
	}

	lockTx := wire.NewMsgTx(wire.TxVersion)
	err = lockTx.Deserialize(bytes.NewReader(tracker.ProcessUnsignedTx))
	if err != nil {
		ctx.Logger.Error("err trying to deserialize btc txn: ", err, j.TrackerName)
		return
	}

	type sign []byte
	btcSignatures := tracker.Multisig.GetSignatures()
	signatures := make([]sign, len(btcSignatures))
	for i := range btcSignatures {
		index := btcSignatures[i].Index
		signatures[index] = btcSignatures[i].Sign
	}

	builder := txscript.NewScriptBuilder().AddOp(txscript.OP_FALSE)
	for i := range signatures {
		builder.AddData(signatures[i])
		if i == tracker.Multisig.M {
			break
		}
	}

	lockScript, err := ctx.LockScripts.GetLockScript(tracker.CurrentLockScriptAddress)
	if err != nil {
		ctx.Logger.Error("err trying to get lockscript ", err, j.TrackerName)
		return
	}

	builder.AddFullData(lockScript)
	sigScript, err := builder.Script()

	cd := bitcoin.NewChainDriver(ctx.BlockCypherToken)
	lockTx = cd.AddLockSignature(tracker.ProcessUnsignedTx, sigScript)

	buf := bytes.NewBuffer([]byte{})
	err = lockTx.Serialize(buf)
	if err != nil {
		ctx.Logger.Error("err trying to serialize btc final txn ", err, j.TrackerName)
		return
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         ctx.BTCNodeAddress + ":" + ctx.BTCRPCPort,
		User:         ctx.BTCRPCUsername,
		Pass:         ctx.BTCRPCPassword,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	clt, err := rpcclient.New(connCfg, nil)
	if err != nil {
		ctx.Logger.Error("err trying to connect to bitcoin node", j.TrackerName)
		return
	}

	var txBytes []byte
	buf = bytes.NewBuffer(txBytes)
	lockTx.Serialize(buf)
	txBytes = buf.Bytes()

	ctx.Logger.Debug(hex.EncodeToString(txBytes))

	hash, err := cd.BroadcastTx(lockTx, clt)
	if err == nil {

		ctx.Logger.Info("bitcoin tx successful", hash)

		bs := btc.BroadcastSuccess{
			tracker.Name,
			ctx.ValidatorAddress,
			*hash,
		}

		txData, err := bs.Marshal()
		if err != nil {
			ctx.Logger.Error("error while preparing mint txn ", err, j.TrackerName)
			return
		}
		tx := action.RawTx{
			Type: action.BTC_BROADCAST_SUCCESS,
			Data: txData,
			Fee:  action.Fee{},
			Memo: j.JobID,
		}

		req := InternalBroadcastRequest{
			RawTx: tx,
		}
		rep := BroadcastReply{}

		err = ctx.Service.InternalBroadcast(req, &rep)
		if err != nil {
			ctx.Logger.Error("error while broadcasting finality vote and mint txn ", err, j.TrackerName)
			return
		}

	} else {
		ctx.Logger.Error("broadcast failed err: ", err, " tracker: ", j.TrackerName)
	}

}

/*
	simple getters
*/
func (j *JobBTCBroadcast) GetType() string {
	return JobTypeBTCBroadcast
}

func (j *JobBTCBroadcast) GetJobID() string {
	return j.JobID
}

func (j JobBTCBroadcast) IsDone() bool {
	return j.Status == jobs.Completed
}
