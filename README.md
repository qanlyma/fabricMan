# FabricMan

## 主要修改

1. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go` 在 ReadSet 中加入了 Value
2. `core/endorser/endorser.go` 如果第一个参数是 `transfer` 标记转账交易
3. `orderer/common/blockcutter/scheduler/scheduler.go` 用于合并交易（合并的交易标为 0，被合并的交易标为 1）
4. `fabric-protos-go/common/common.pb.go` 在 Envelope 中加入合并交易的标识和 payload
5. `core/ledger/kvledger/txmgmt/validation/batch_preparer.go` 将合并交易写集加入 updates