package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"goblockchain/transaction"
	"goblockchain/utils"
	"time"
)

// 我们其实可以认为，Block(区块)相当于链表中的一个节点
type Block struct {
	Timestamp    int64                      // 当前时间戳
	Hash         []byte                     // 当前区块的哈希值
	PrevHash     []byte                     // 存储着前一个区块的哈希值
	Target       []byte                     // 目标难度值
	Nonce        int64                      // 随机值
	Transactions []*transaction.Transaction // 区块中的交易信息
}

// 将Transactions中的ID序列转化为[]byte类型，方便SetHash函数融合Block的属性
func (b *Block) BackTrasactionSummary() []byte {
	txIDs := make([][]byte, 0)
	// range遍历[]*transaction.Transaction切片，将交易ID都存放进txIDs数组中
	for _, tx := range b.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	summary := bytes.Join(txIDs, []byte{})
	return summary
}

// 得到当前区块的哈希值
func (b *Block) SetHash() {
	// bytes.Join可以连接多个字节串，第二个参数为将字节串连接时的分隔符。此处[]byte{}即为空
	information := bytes.Join([][]byte{utils.ToHexInt(b.Timestamp), b.PrevHash, b.Target, utils.ToHexInt(b.Nonce), b.BackTrasactionSummary()}, []byte{})
	hash := sha256.Sum256(information)
	// 区块的哈希值相当于其ID值，同时也可以用来检查区块所包含信息的完整性
	b.Hash = hash[:]
}

// 创建区块
func CreateBlock(prevhash []byte, txs []*transaction.Transaction) *Block {
	block := Block{time.Now().Unix(), []byte{}, prevhash, []byte{}, 0, txs}
	// 更新区块的Target和Nonce属性
	block.Target = block.GetTarget()
	block.Nonce = block.FindNonce()
	// 计算总的哈希值并更新区块的Hash属性
	block.SetHash()
	return &block
}

// 我们还需要一个初始区块，以防blockchain.go中的AddBlock获取前一个区块出错
func GenesisBlock(address []byte) *Block {
	tx := transaction.BaseTx(address)
	genesis := CreateBlock([]byte("Lei is awesome!"), []*transaction.Transaction{tx})
	genesis.SetHash()
	return genesis
}

// 序列化区块生成字节串的
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	utils.Handle(err)
	return res.Bytes()
}

// 反序列化区块
func DeSerializeBlock(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	utils.Handle(err)
	return &block
}
