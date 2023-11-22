# FabricMan

说明：本项目由我独立完成，提交者有两个是因为主要代码在实验室的服务器上，使用了其他同学账号进行提交。

## 主要修改

### 交易合并：

1. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_builder.go` 在 ReadSet 中加入了 Value
2. `core/ledger/kvledger/txmgmt/rwsetutil/rwset_proto_util.go` 在 TxRwSet 中加入 MergeSign `res.MergeSign = []byte{'1'}`
3. `vendor/github.com/hyperledger/fabric-protos-go/common/common.pb.go` 在 Envelope 中加入合并交易的标识 MergeSign（可去掉？） 和 MergePayload（传递合并的读写集）
4. `vendor/github.com/hyperledger/fabric-protos-go/peer/transaction.pb.go` 加入 TxValidationCode 通知被合并的交易

5. `core/endorser/endorser.go` 如果符合交易规则，TxRwSet.MergeSign 标记为转账交易
6. `orderer/common/blockcutter/scheduler/scheduler.go` 合并交易（合并的交易 MergeSign 标为 0），并将合并后的读写集传给 peer
7. `core/ledger/kvledger/txmgmt/validation/batch_preparer.go` 将合并交易读写集加入 internalBlock
8. `core/ledger/kvledger/txmgmt/validation/validator.go` 只对合并交易的读写集进行 MVCC 验证，并对被合并的交易直接返回相应 TxValidationCode
 
### 交易重排：

fabric++ 中的交易重排：

#### tarjanscc.go

实现了 Tarjan 算法来寻找图中的强连通分量（Strongly Connected Components，SCC）：

1. 导入所需的包。
2. 定义了一个自定义类型 `ById`，用于对 int32 类型的值进行排序。
3. `min` 函数返回两个 int32 数字的较小值。
4. `SCC` 结构体表示一个强连通分量，包含两个字段：Vertices（int32 的切片）用于存储该分量中的顶点，Member（bool 的切片）用于表示顶点是否属于该分量。
5. `TarjanSCC` 接口定义了三个方法：`SCC()` 用于计算 SCC 的数量，`GetSCCs()` 用于获取 SCC 的列表，`SCCUtil(u int32)` 用于执行递归的深度优先搜索遍历以寻找 SCC。
6. `tarjanscc` 结构体实现了 `TarjanSCC` 接口，并保存了 Tarjan 算法所需的数据。
7. `NewTarjanSCC` 函数使用提供的邻接矩阵创建了 `tarjanscc` 的新实例，并初始化各种数据结构。
8. `SCC()` 方法是启动 SCC 计算的驱动方法。它将所有顶点标记为未访问，并对每个未访问的顶点调用 `SCCUtil()` 方法。
9. `SCCUtil()` 方法是一个递归的深度优先搜索（DFS）遍历，用于查找 SCC。它为顶点分配发现时间和 low 值，将顶点入栈，并遍历相邻顶点。它根据相邻顶点的 low 值更新顶点的 low 值。如果顶点形成一个强连通分量，它将从栈中弹出，并将该 SCC 存储在 sccList 中。
10. `GetSCCs()` 方法返回强连通分量的列表。

#### johnsonce.go

实现了 Johnson 算法，对给定的有向图进行强连通分量分析，并处理其中的循环，最终返回移除的顶点数量和无效顶点的布尔数组：

1. 该代码定义了一个 `JohnsonCE` 接口和一个 `johnsonce` 结构体，其中 `johnsonce` 实现了 `JohnsonCE` 接口的方法。
2. `Run()` 方法用于运行 Johnson's algorithm，找到并处理强连通分量中的循环。它返回移除的顶点数量和一个布尔数组，表示哪些顶点被标记为无效。
3. `FindCycles()` 方法用于在给定的强连通分量中找到所有的循环。它返回两个矩阵，一个表示顶点是否属于循环，另一个表示每个顶点属于的循环数量。
4. `FindCyclesRecur()` 方法是递归方法，用于在当前顶点的邻居中查找循环。
5. `BreakCycles()` 方法用于处理循环。它根据循环的数量和顶点的权重来确定要移除的顶点，然后将其标记为无效。

fabricMan 中的交易重排：

#### reorderSort.go

由于考虑到 fabric++ 中的交易重排在环路多的时候算法时间复杂度过高，此处重新设计了一个简单的交易重排算法。

### 并行验证：

1. 在 reorder 的过程中划分连通子图，不在同一个子图内的交易相互独立可以并行验证。
2. `vendor/github.com/hyperledger/fabric-protos-go/common/common.pb.go` 在 Envelope 中加入 Subgraphs 表示相关联的交易。
3. `core/ledger/kvledger/txmgmt/validation/batch_preparer.go` 接受 Subgraphs。
4. `core/ledger/kvledger/txmgmt/validation/validator.go` 通过 Subgraphs 来并行处理交易。
