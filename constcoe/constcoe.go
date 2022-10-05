package constcoe

// 实现PoW时的难度
// 对于256位的哈希值来说设定挖矿条件的方式往往是：前多少位为0。
// Difficulty便是用于指定目标哈希需满足的条件的，即计算的哈希值必须前Difficulty位为0.
// https://github.com/chaors/PublicBlockChain_go/blob/master/part2-ProofOfWork-Prototype/goLang%E5%85%AC%E9%93%BE%E5%AE%9E%E6%88%98%E4%B9%8BProofOfWork.md
const (
	Difficulty          = 12
	InitCoin            = 1000                          // 代表了区块链在创建时的总的比特币数目
	TransactionPoolFile = "./tmp/transaction_pool.data" // 临时交易池
	BCPath              = "./tmp/blocks"                //
	BCFile              = "./tmp/blocks/MANIFEST"       //
)
