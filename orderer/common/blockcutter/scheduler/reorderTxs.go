// fabricMan

package scheduler

import (
	"time"

	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
)

const maxUniqueKeys = 65563

var scheduler TxnScheduler
var pendingBatch map[int]*cb.Envelope

type TxnScheduler struct {
	maxTxnCount   uint32
	invalid       []bool
	keyVersionMap map[uint32]*kvrwset.Version
	keyTxMap      map[uint32][]int32

	txReadSet  [][]uint64
	txWriteSet [][]uint64

	uniqueKeyCounter uint32
	uniqueKeyMap     map[string]uint32
	pendingTxns      []int
}

func NewTxnScheduler(blkSize uint32) TxnScheduler {

	return TxnScheduler{
		maxTxnCount: blkSize,

		txReadSet:  make([][]uint64, blkSize),
		txWriteSet: make([][]uint64, blkSize),

		invalid:       make([]bool, blkSize),
		keyVersionMap: make(map[uint32]*kvrwset.Version),
		keyTxMap:      make(map[uint32][]int32),

		uniqueKeyCounter: 0,
		uniqueKeyMap:     make(map[string]uint32),

		pendingTxns: make([]int, 0),
	}
}

func reorderBatch() []int {
	var validCount, invalidCount int
	defer func(start time.Time) {
		elapsed := time.Since(start).Nanoseconds() / 1000
		logger.Infof("Process Blk in %d us ( %d valid txns, %d invalid txns)", elapsed, validCount, invalidCount)
	}(time.Now())
	if len(scheduler.pendingTxns) <= 1 {
		return scheduler.pendingTxns
	}

	txnCount := len(scheduler.pendingTxns)
	graph := make([][]int32, txnCount)
	invgraph := make([][]int32, txnCount)
	for i := int32(0); i < int32(txnCount); i++ {
		graph[i] = make([]int32, 0, txnCount)
		invgraph[i] = make([]int32, 0, txnCount)
	}

	// for every transactions, find the intersection between the readSet and the writeSet
	start := time.Now()
	for i := int32(0); i < int32(txnCount); i++ {
		for j := int32(0); j < int32(txnCount); j++ {
			if i == j || scheduler.invalid[i] || scheduler.invalid[j] {
				continue
			} else {
				for k := uint32(0); k < (maxUniqueKeys / 64); k++ {
					if (scheduler.txWriteSet[i][k] & scheduler.txReadSet[j][k]) != 0 {
						// Txn j must be scheduled before txn i
						graph[i] = append(graph[i], j)
						invgraph[j] = append(invgraph[j], i)
						break
					}
				}
			}
		}
	}
	elapsedDependency := time.Since(start).Nanoseconds() / 1000
	logger.Infof("Resolve in-blk txn dependency in %d us", elapsedDependency)

	start = time.Now()
	resGen := NewResolver(&graph, &invgraph)

	res, _ := resGen.GetSchedule()
	lenres := len(res)
	elapsedSchedule := time.Since(start).Nanoseconds() / 1000
	logger.Infof("Schedule txns in %d ", elapsedSchedule)

	resGen = nil
	graph = nil
	invgraph = nil

	validBatch := make([]int, 0)

	for i := 0; i < lenres; i++ {
		validBatch = append(validBatch, scheduler.pendingTxns[res[lenres-1-i]])
	}

	validCount = lenres
	invalidCount = 0
	for _, valid := range scheduler.invalid {
		if valid {
			invalidCount++
		}
	}
	// log some information
	logger.Infof("schedule len %d -> %v", len(res), res)
	logger.Infof("Finish processing blk ")
	return validBatch
}
