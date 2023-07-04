// fabricMan

package scheduler

import (
	"encoding/json"
	"strconv"
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	utils "github.com/hyperledger/fabric/protoutil"
)

var logger = flogging.MustGetLogger("orderer.common.blockcutter.scheduler")

func ScheduleTxn(batch []*cb.Envelope) []*cb.Envelope {
	logger.Info("============================================================>>> 2.4 ScheduleTxn!!!")

	initStrcut(len(batch))
	unMarshalAndSort(batch)

	if len(batch) < 2 && len(transferSet) == 0 {
		return batch
	}

	newbatch := make([]*cb.Envelope, 0)
	newSubs := make([][]int, 0)
	orderMap := make(map[int]int)

	// merge
	if len(transferSet) > 0 {
		mergeTransferTxs(batch)
	}

	// reorder
	schedule, subs := reorderBatch()
	logger.Info("subs:", subs)

	// build new batch
	for i, txnID := range schedule {
		logger.Info("schedule ordering: ", i, txnID)
		orderMap[txnID] = i
		newbatch = append(newbatch, pendingBatch[txnID])
	}

	// pvalidation
	logger.Info("ordermap: ", orderMap)
	for i := 0; i < len(subs); i++ {
		sub := subs[i]
		temp := make([]int, 0)
		for j := 0; j < len(sub); j++ {
			temp = append(temp, orderMap[int(sub[j])])
		}
		newSubs = append(newSubs, temp)
	}
	subData, err := json.Marshal(newSubs)
	if err != nil {
		logger.Info("err")
	}
	newbatch[0].Subgraphs = subData

	logger.Info("============================================================>>> newbatch:", newSubs)
	return newbatch
}

func initStrcut(size int) {
	transferSet = make([]Transfer, 0)
	moneyMap = make(map[string]int)
	versionMap = make(map[string]*kvrwset.Version)
	pendingBatch = make(map[int]*cb.Envelope)
	scheduler = NewTxnScheduler(uint32(size))
}

func unMarshalAndSort(batch []*cb.Envelope) {
	logger.Info("=======================================================>>> Received txRWSet!!!")

	for i, msg := range batch {
		logger.Infof("|||||||||||||||||| Tx %d:", i)

		// unMarshal
		resppayload, err := utils.GetActionFromEnvelopeMsg(msg)
		if err != nil {
			logger.Info("err 1")
		}
		txRWSet := &rwsetutil.TxRwSet{}
		err = txRWSet.FromProtoBytes(resppayload.Results)
		if err != nil {
			logger.Info("err 2")
		}
		logger.Info("is transferm:", txRWSet.MergeSign != nil)

		ns := txRWSet.NsRwSets[1]
		printTxRWSet(ns)
		readSet := make([]uint64, maxUniqueKeys/64)
		writeSet := make([]uint64, maxUniqueKeys/64)
		tid := int32(len(scheduler.pendingTxns))

		// sort
		if txRWSet.MergeSign != nil {
			// merge part
			var fr, to, money int
			contract = ns.NameSpace
			moneystr, _ := strconv.Atoi(string(ns.KvRwSet.Reads[1].GetValue()))
			moneyend, _ := strconv.Atoi(string(ns.KvRwSet.Writes[0].GetValue()))

			if moneystr > moneyend {
				fr = 1
				to = 2
				money = moneystr - moneyend
			} else {
				fr = 2
				to = 1
				money = moneyend - moneystr
			}
			logger.Infof("start: %d end: %d fr: %d to %d", moneystr, moneyend, fr, to)

			// add Tx to list
			f := ns.KvRwSet.Reads[fr].GetKey()
			t := ns.KvRwSet.Reads[to].GetKey()
			transferSet = append(transferSet, Transfer{tx: msg, from: f, to: t, val: money})

			// add money to moneyMap, version to versionMap
			verfr := ns.KvRwSet.Reads[fr].GetVersion()
			verto := ns.KvRwSet.Reads[to].GetVersion()
			moneyfr, _ := strconv.Atoi(string(ns.KvRwSet.Reads[fr].GetValue()))
			moneyto, _ := strconv.Atoi(string(ns.KvRwSet.Reads[to].GetValue()))
			if _, exist := moneyMap[f]; !exist {
				moneyMap[f] = moneyfr
				versionMap[f] = verfr
			}
			if _, exist := moneyMap[t]; !exist {
				moneyMap[t] = moneyto
				versionMap[t] = verto
			}

		} else {
			// reorder part
			readKeys := []string{}
			writeKeys := []string{}
			defer func(start time.Time) {
				elapsed := time.Since(start).Nanoseconds() / 1000
				logger.Infof("Process txn with read keys %v and write keys %v in %d us", readKeys, writeKeys, elapsed)
			}(time.Now())

			for _, write := range ns.KvRwSet.Writes {
				if writeKey := write.GetKey(); validKey(writeKey) {
					writeKeys = append(writeKeys, writeKey)

					// check if the key exists
					key, ok := scheduler.uniqueKeyMap[writeKey]

					if !ok {
						// if the key is not found, insert and increment
						// the key counter
						scheduler.uniqueKeyMap[writeKey] = scheduler.uniqueKeyCounter
						key = scheduler.uniqueKeyCounter
						scheduler.uniqueKeyCounter += 1
					}

					// set the respective bit in the writeSet
					index := key / 64
					writeSet[index] |= (uint64(1) << (key % 64))
				}
			}

			for _, read := range ns.KvRwSet.Reads {
				if readKey := read.GetKey(); validKey(readKey) {
					readVer := read.GetVersion()
					readKeys = append(readKeys, readKey)

					key, ok := scheduler.uniqueKeyMap[readKey]
					if !ok {
						// if the key is not found, it is inserted. So increment
						// the key counter
						scheduler.uniqueKeyMap[readKey] = scheduler.uniqueKeyCounter
						key = scheduler.uniqueKeyCounter
						scheduler.uniqueKeyCounter += 1
					}

					ver, ok := scheduler.keyVersionMap[key]
					if ok {
						if ver.BlockNum == readVer.BlockNum && ver.TxNum == readVer.TxNum {
							scheduler.keyTxMap[key] = append(scheduler.keyTxMap[key], tid)
						} else {
							// It seems to abort the previous txns with for the unmatched version
							// logger.Infof("Invalidate txn %v", r.keyTxMap[key])
							for _, tx := range scheduler.keyTxMap[key] {
								scheduler.invalid[tx] = true
							}
							scheduler.keyTxMap[key] = nil
						}
					} else {
						scheduler.keyTxMap[key] = append(scheduler.keyTxMap[key], tid)
						scheduler.keyVersionMap[key] = readVer
					}

					index := key / 64
					readSet[index] |= (uint64(1) << (key % 64))
				}
			}
		}

		scheduler.txReadSet[tid] = readSet
		scheduler.txWriteSet[tid] = writeSet
		scheduler.pendingTxns = append(scheduler.pendingTxns, i)
		pendingBatch[i] = msg
	}
	logger.Info("=======================================================>>> End of txRWSet!!!")
}
