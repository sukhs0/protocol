package event

import (

	"github.com/Oneledger/protocol/config"
	"github.com/Oneledger/protocol/data/jobs"
	"github.com/Oneledger/protocol/log"
	"os"
	"github.com/Oneledger/protocol/chains/ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

var _ jobs.Job = &JobETHBroadcast{}
type JobETHBroadcast struct {
	TrackerName         ethereum.TrackerName
	JobID               string
	RetryCount          int
	JobStatus          	jobs.Status
}

func (job JobETHBroadcast) DoMyJob(ctx interface{}) {

	// get tracker
	job.RetryCount += 1
	if job.RetryCount > jobs.Max_Retry_Count{
		job.JobStatus = jobs.Failed
	}
	if job.JobStatus == jobs.New {
		job.JobStatus = jobs.InProgress
	}
	ethCtx, _ := ctx.(*JobsContext)
	trackerStore := ethCtx.EthereumTrackers
	tracker, err := trackerStore.Get(job.TrackerName)
	if err != nil {
		ethCtx.Logger.Error("err trying to deserialize tracker: ", job.TrackerName, err)
		return
	}
	ethconfig := config.DefaultEthConfig()
	logger := log.NewLoggerWithPrefix(os.Stdout,"JOB_ETHBROADCAST")
	cd,err := ethereum.NewEthereumChainDriver(ethconfig,logger,&ethCtx.ETHPrivKey)
	if err != nil {
		ethCtx.Logger.Error("err trying to get ChainDriver : ", job.GetJobID(), err)

		return
	}
	rawTx := tracker.SignedETHTx
	tx := &types.Transaction{}
	err = rlp.DecodeBytes(rawTx, tx)
	if err != nil {
		ethCtx.Logger.Error("Error Decoding Bytes from RaxTX :", job.GetJobID(),err)
		return
	}
	_,err = cd.BroadcastTx(tx)
	if err != nil {
		ethCtx.Logger.Error("Error in transaction broadcast : ", job.GetJobID(),err)
		return
	}
	job.JobStatus = jobs.Completed
}

func (job JobETHBroadcast) IsMyJobDone(ctx interface{}) bool {

	panic("implement me")
}

func (job JobETHBroadcast) IsSufficient(ctx interface{}) bool {
	panic("implement me")
}

func (job JobETHBroadcast) DoFinalize() {
	panic("implement me")
}

func (job JobETHBroadcast) GetType() string {
	panic("implement me")
}

func (job JobETHBroadcast) GetJobID() string {
	return "We should make a Job ID"
}

func (job JobETHBroadcast) IsDone() bool {
	if job.JobStatus==jobs.Completed {
		return true
	}
	return false
}

