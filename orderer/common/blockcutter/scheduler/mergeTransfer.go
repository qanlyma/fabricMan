// fabricMan

package scheduler

import (
	"strconv"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
)

var transferSet []Transfer
var moneyMap map[string]int
var versionMap map[string]*kvrwset.Version
var contract string

type Transfer struct {
	tx   *cb.Envelope
	from string
	to   string
	val  int
}

func mergeTransferTxs(batch []*cb.Envelope) *cb.Envelope {
	logger.Infof("Numbers of transfer: %d, %+v, %+v", len(transferSet), transferSet, moneyMap)
	return buildMergeMsg(transferSet[0].tx)
}

func buildMergeMsg(base *cb.Envelope) *cb.Envelope {
	logger.Info("buildMergeMsg...")
	for _, t := range transferSet {
		if moneyMap[t.from] >= t.val {
			moneyMap[t.from] -= t.val
			moneyMap[t.to] += t.val
		}
	}
	logger.Info("moneyMap after building: ", moneyMap)

	var ws []*kvrwset.KVWrite
	for k, v := range moneyMap {
		kv := &kvrwset.KVWrite{Key: k, Value: []byte(strconv.Itoa(v))}
		ws = append(ws, kv)
	}
	var rs []*kvrwset.KVRead
	for k, v := range versionMap {
		kv := &kvrwset.KVRead{Key: k, Version: v}
		rs = append(rs, kv)
	}

	rws := &kvrwset.KVRWSet{Reads: rs, Writes: ws}
	ns := &rwsetutil.NsRwSet{NameSpace: contract, KvRwSet: rws}
	printTxRWSet(ns)
	nss := make([]*rwsetutil.NsRwSet, 1)
	nss[0] = ns
	txRWSet := &rwsetutil.TxRwSet{NsRwSets: nss, MergeSign: []byte{'0'}}

	// Add WriteSet in Envelope.MergePayload of 'base' for simplicity
	pl, _ := txRWSet.ToProtoBytes()
	base.MergePayload = pl
	base.MergeSign = []byte{'0'}
	return base
}
