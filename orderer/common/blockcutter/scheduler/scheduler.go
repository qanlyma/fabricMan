package scheduler

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	utils "github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("orderer.common.blockcutter")

func ScheduleTxn(batch []*cb.Envelope) []*cb.Envelope {
	logger.Info("======================================================================>>> 2.6 ScheduleTxn!!!")
	logger.Info("======================================================================>>> Received txRWSet!!!")

	for i, msg := range batch {
		logger.Infof("Tx %d:", i+1)
		resppayload, _ := utils.GetActionFromEnvelopeMsg(msg)
		txRWSet := &rwsetutil.TxRwSet{}
		_ = txRWSet.FromProtoBytes(resppayload.Results)
		ns := txRWSet.NsRwSets[1]

		for _, read := range ns.KvRwSet.Reads {
			v := "nil"
			if read.GetValue() != nil {
				v = string(read.GetValue())
			}
			if read.GetVersion() == nil {
				logger.Infof("Read Key: %s, Version: nil, Value: %s", read.GetKey(), v)
			} else {
				logger.Infof("Read Key: %s, Version: (%d, %d), Value: %s", read.GetKey(), read.GetVersion().GetBlockNum(), read.GetVersion().GetTxNum(), v)
			}
		}
		for _, write := range ns.KvRwSet.Writes {
			logger.Infof("Write Key: %s, Value: %s", write.GetKey(), string(write.GetValue()))
		}
	}

	logger.Infof("======================================================================>>> End of txRWSet!!!")
	return batch
}
