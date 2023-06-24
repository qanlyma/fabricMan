// fabricMan

package scheduler

import (
	"strconv"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	utils "github.com/hyperledger/fabric/protoutil"
)

type Transfer struct {
	tx   *cb.Envelope
	from string
	to   string
	val  int
}

var transferSet []Transfer
var moneyMap map[string]int
var versionMap map[string]*kvrwset.Version
var contract string

var logger = flogging.MustGetLogger("orderer.common.blockcutter.scheduler")

func ScheduleTxn(batch []*cb.Envelope) []*cb.Envelope {
	logger.Info("============================================================>>> 2.4 ScheduleTxn!!!")

	batch = mergeTransferTxs(batch)

	return batch
}

func unMarshalAndSort(batch []*cb.Envelope) {
	logger.Info("=======================================================>>> Received txRWSet!!!")

	for i, msg := range batch {
		logger.Infof("|||||||||||||||||| Tx %d:", i+1)

		// UnMarshal
		resppayload, err := utils.GetActionFromEnvelopeMsg(msg)
		if err != nil {
			logger.Info("err 1")
		}
		txRWSet := &rwsetutil.TxRwSet{}
		err = txRWSet.FromProtoBytes(resppayload.Results)
		if err != nil {
			logger.Info("err 2")
		}
		logger.Info("is transfer:", txRWSet.MergeSign != nil)

		ns := txRWSet.NsRwSets[1]
		contract = ns.NameSpace
		printTxRWSet(ns)

		// Sort transferTxs
		if txRWSet.MergeSign != nil {
			// Add MergeSign to Envelope
			msg.MergeSign = []byte{'1'}
			// Add Tx to list
			money, _ := strconv.Atoi(string(txRWSet.MergeSign))
			a := ns.KvRwSet.Reads[1].GetKey()
			b := ns.KvRwSet.Reads[2].GetKey()
			transferSet = append(transferSet, Transfer{tx: msg, from: a, to: b, val: money})
			// Add money to moneyMap
			money1, _ := strconv.Atoi(string(ns.KvRwSet.Reads[1].GetValue()))
			money2, _ := strconv.Atoi(string(ns.KvRwSet.Reads[2].GetValue()))
			ver1 := ns.KvRwSet.Reads[1].GetVersion()
			ver2 := ns.KvRwSet.Reads[2].GetVersion()
			if _, exist := moneyMap[a]; !exist {
				moneyMap[a] = money1
				versionMap[a] = ver1
			}
			if _, exist := moneyMap[b]; !exist {
				moneyMap[b] = money2
				versionMap[b] = ver2
			}
		}
	}
	logger.Info("=======================================================>>> End of txRWSet!!!")
}
