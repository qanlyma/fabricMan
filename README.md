# FabricMan

## 主要修改

### 交易合并：

1. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go` 在 ReadSet 中加入了 Value
2. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go` 在 TxRwSet 中加入 MergeSign（传递转账金额）
3. `fabric-protos-go/common/common.pb.go` 在 Envelope 中加入合并交易的标识 MergeSign 和 MergePayload（传递合并的读写集）
4. `vendor/github.com/hyperledger/fabric-protos-go/peer/transaction.pb.go` 加入 TxValidationCode

4. `core/endorser/endorser.go` 如果第一个参数是 `transfer`，TxRwSet.MergeSign 标记为转账交易
5. `orderer/common/blockcutter/scheduler/scheduler.go` 用于合并交易（合并的交易标为 0，被合并的交易标为 1），并将合并后的读写集传给 peer
6. `core/ledger/kvledger/txmgmt/validation/batch_preparer.go` 将合并交易写集加入 internalBlock

### 交易重排：