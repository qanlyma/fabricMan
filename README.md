# FabricMan

## 主要修改

### 交易合并：

1. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go` 在 ReadSet 中加入了 Value
2. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go` 在 TxRwSet 中加入 MergeSign `res.MergeSign = []byte{'1'}`
3. `fabric-protos-go/common/common.pb.go` 在 Envelope 中加入合并交易的标识 MergeSign（可去掉？） 和 MergePayload（传递合并的读写集）
4. `vendor/github.com/hyperledger/fabric-protos-go/peer/transaction.pb.go` 加入 TxValidationCode 通知被合并的交易

5. `core/endorser/endorser.go` 如果第一个参数是 `transfer`，TxRwSet.MergeSign 标记为转账交易
6. `orderer/common/blockcutter/scheduler/scheduler.go` 合并交易（合并的交易 MergeSign 标为 0），并将合并后的读写集传给 peer
7. `core/ledger/kvledger/txmgmt/validation/batch_preparer.go` 将合并交易读写集加入 internalBlock
8. `core/ledger/kvledger/txmgmt/validation/validator.go` 只对合并交易的读写集进行 MVCC 验证，并对被合并的交易直接返回相应 TxValidationCode
 
### 交易重排：

#### tarjanscc.go

实现了 Tarjan 算法来寻找图中的强连通分量（Strongly Connected Components，SCC）：

1. 导入所需的包。
2. 定义了一个自定义类型 ById，用于对 int32 类型的值进行排序。
3. min 函数返回两个 int32 数字的较小值。
4. SCC 结构体表示一个强连通分量，包含两个字段：Vertices（int32 的切片）用于存储该分量中的顶点，Member（bool 的切片）用于表示顶点是否属于该分量。
5. TarjanSCC 接口定义了三个方法：SCC() 用于计算 SCC 的数量，GetSCCs() 用于获取 SCC 的列表，SCCUtil(u int32) 用于执行递归的深度优先搜索遍历以寻找 SCC。
6. tarjanscc 结构体实现了 TarjanSCC 接口，并保存了 Tarjan 算法所需的数据。
7. NewTarjanSCC 函数使用提供的邻接矩阵创建了 tarjanscc 的新实例，并初始化各种数据结构。
8. SCC() 方法是启动 SCC 计算的驱动方法。它将所有顶点标记为未访问，并对每个未访问的顶点调用 SCCUtil() 方法。
9. SCCUtil() 方法是一个递归的深度优先搜索（DFS）遍历，用于查找 SCC。它为顶点分配发现时间和 low 值，将顶点入栈，并遍历相邻顶点。它根据相邻顶点的 low 值更新顶点的 low 值。如果顶点形成一个强连通分量，它将从栈中弹出，并将该 SCC 存储在 sccList 中。
10. GetSCCs() 方法返回强连通分量的列表。

#### johnsonce.go