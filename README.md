# FabricMan

因为是用服务器提交的代码，所以不是我自己的 ID。

## 主要修改

1. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go` 在 ReadSet 中加入了 Value
2. `orderer/common/blockcutter/scheduler/scheduler.go` 用于合并交易
3. `fabric-protos-go/common/common.pb.go` 在 Envelope 中加入合并交易的标识
4. `core/endorser/endorser.go` 如果第一个参数是 `transfer` 标记转账交易