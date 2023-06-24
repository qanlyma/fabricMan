// fabricMan

package scheduler

import (
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
)

func printTxRWSet(ns *rwsetutil.NsRwSet) {
	logger.Infof("Contract: %s", ns.NameSpace)
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
