package blockchain

import (
	"bytes"
	"crypto/sha256"
	"goblockchain/constcoe"
	"goblockchain/utils"
	"math"
	"math/big"
)

/**
* 共识机制说的通俗明白一点就是要在相对公平的条件下让想要添加区块进区块链的节点内卷，
* 通过竞争选择出一个大家公认的节点添加它的区块进入区块链。整个共识机制被分为两部分，首先是竞争，然后是共识。
* 中本聪在比特币中设计了如下的一个Game来实现竞争：
* 每个节点去寻找一个随机值（也就是nonce），将这个随机值作为候选区块的头部信息属性之一，
* 要求候选区块对自身信息（注意这里是包含了nonce的）进行哈希后表示为数值要小于一个难度目标值（也就是Target），
* 最先寻找到nonce的节点即为卷王，可以将自己的候选区块发布并添加到区块链尾部。
* 这个Game设计的非常巧妙，首先每个节点要寻找到的nonce只对自己候选区块有效，防止了其它节点同学抄答案；
* 其次，nonce的寻找是完全随机的没有技巧，寻找到nonce的时间与目标难度值与节点本身计算性能有关，但不妨碍性能较差的节点也有机会获胜；
* 最后寻找nonce可能耗费大量时间与资源，但是验证卷王是否真的找到了nonce却非常却能够很快完成并几乎不需要耗费资源，
* 这个寻找到的nonce可以说就是卷王真的是卷王的证据。
 */

// 返回目标难度值。
// 我们这里使用的之前设定的一个常量Difficulty来构造目标难度值
// 但是在实际的区块链中目标难度值会根据网络情况定时进行调整，且能够保证各节点在同一时间在同一难度下进行竞争
// 故这里的GetTarget可以理解为预留API，期待一下之后的分布式网络实现。
func (b *Block) GetTarget() []byte {
	// 创建一个初始值为1的target
	target := big.NewInt(1)
	// target左移256 - 12 = 244位
	target.Lsh(target, uint(256-constcoe.Difficulty))
	return target.Bytes()
}

// 拼接区块属性，返回字节数组
func (b *Block) GetBase4Nonce(nonce int64) []byte {
	data := bytes.Join([][]byte{
		utils.ToHexInt(b.Timestamp),
		b.PrevHash,
		b.Target,
		utils.ToHexInt(int64(nonce)),
		b.BackTrasactionSummary(),
	},
		[]byte{},
	)
	return data
}

// 针对某一区块计算符合要求的nonce值
func (b *Block) FindNonce() int64 {
	var intHash big.Int
	var intTarget big.Int
	var hash [32]byte
	var nonce int64
	nonce = 0
	// 将b.Target转化为big.Int类型赋值给intTarget
	intTarget.SetBytes(b.Target)

	for nonce < math.MaxInt64 {
		// 合并该区块的所有属性，返回字节数组
		data := b.GetBase4Nonce(nonce)
		// 计算该区块的哈希值
		hash = sha256.Sum256(data)
		intHash.SetBytes(hash[:])
		// 当哈希值intHash小于边界值intTarget则结束循环，否则nonce+1继续计算哈希值
		if intHash.Cmp(&intTarget) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce
}

// 验证区块是否符合PoW要求
func (b *Block) ValidatePoW() bool {
	var intHash big.Int
	var intTarget big.Int
	var hash [32]byte

	intTarget.SetBytes(b.Target)
	data := b.GetBase4Nonce(b.Nonce)
	hash = sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	if intHash.Cmp(&intTarget) == -1 {
		return true
	}
	return false
}
