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

var logger = flogging.MustGetLogger("orderer.common.blockcutter.scheduler")

func ScheduleTxn(batch []*cb.Envelope) []*cb.Envelope {
	logger.Info("============================================================>>> 2.4 ScheduleTxn!!!")

	moneyMap = make(map[string]int)
	transferSet = unMarshalAndSort(batch)
	logger.Infof("Numbers of transfer: %d, %+v, %+v", len(transferSet), transferSet, moneyMap)

	if len(transferSet) > 0 {
		buildMergeMsg(transferSet[0].tx)
	}

	logger.Infof("After buliding: %+v", moneyMap)
	logger.Info("============================================================>>> Complete, len of batch:", len(batch))
	return batch
}

func unMarshalAndSort(batch []*cb.Envelope) (transferSet []Transfer) {
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
		printTxRWSet(ns)

		// Sort
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
			if _, exist := moneyMap[a]; !exist {
				moneyMap[a] = money1
			}
			if _, exist := moneyMap[b]; !exist {
				moneyMap[b] = money2
			}
		}
	}
	logger.Info("=======================================================>>> End of txRWSet!!!")
	return
}

func buildMergeMsg(base *cb.Envelope) *cb.Envelope {
	logger.Info("buildMergeMsg...")
	for _, t := range transferSet {
		if moneyMap[t.from] >= t.val {
			moneyMap[t.from] -= t.val
			moneyMap[t.to] += t.val
		}
	}

	var ws []*kvrwset.KVWrite
	for k, v := range moneyMap {
		kv := &kvrwset.KVWrite{Key: k, Value: []byte(strconv.Itoa(v))}
		ws = append(ws, kv)
	}

	logger.Info("=======================================================>>> ", len(moneyMap), ws)

	rws := &kvrwset.KVRWSet{Writes: ws}
	ns := &rwsetutil.NsRwSet{NameSpace: "simple", KvRwSet: rws}
	printTxRWSet(ns)
	nss := make([]*rwsetutil.NsRwSet, 1)
	nss[0] = ns
	txRWSet := &rwsetutil.TxRwSet{NsRwSets: nss}

	// Add WriteSet in Envelope.MergePayload for simplicity
	pl, _ := txRWSet.ToProtoBytes()
	base.MergePayload = pl
	base.MergeSign = []byte{'0'}
	return base
}

func printTxRWSet(ns *rwsetutil.NsRwSet) {
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
