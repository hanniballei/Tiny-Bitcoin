package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"goblockchain/constcoe"
	"goblockchain/transaction"
	"goblockchain/utils"
	"runtime"

	"github.com/dgraph-io/badger"
)

// 可以看出区块链就是区块的一个集合
type BlockChain struct {
	LastHash []byte     // 指向区块链最后一个区块的哈希值
	Database *badger.DB // 指向存储区块的数据库
}

// 区块链迭代器
type BlockChainIterator struct {
	CurrentHash []byte     //
	Database    *badger.DB // 指向存储区块的数据库
}

// 迭代器初始化函数
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iterator := BlockChainIterator{chain.LastHash, chain.Database}
	return &iterator
}

// 类似Java迭代器中的Next函数
// 迭代器的迭代函数，让每次迭代返回一个block，然后迭代器指向前一个区块的哈希值
func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		utils.Handle(err)

		err = item.Value(func(val []byte) error {
			block = DeSerializeBlock(val)
			return nil
		})
		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	iterator.CurrentHash = block.PrevHash

	return block
}

//blockchain.go
func (chain *BlockChain) BackOgPrevHash() []byte {
	var ogprevhash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("ogprevhash"))
		utils.Handle(err)

		err = item.Value(func(val []byte) error {
			ogprevhash = val
			return nil
		})

		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	return ogprevhash
}

// 区块链添加区块
func (bc *BlockChain) AddBlock(newBlock *Block) {
	var lastHash []byte

	// 提取“lh”键对应的值，即此时区块链的lastHash值
	err := bc.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)

		return err
	})
	utils.Handle(err)
	// 比较新加入的区块的PrevHash值是否与此时区块链的lastHash值一致
	if !bytes.Equal(newBlock.PrevHash, lastHash) {
		fmt.Println("This block is out of age")
		runtime.Goexit()
	}

	err = bc.Database.Update(func(transaction *badger.Txn) error {
		err := transaction.Set(newBlock.Hash, newBlock.Serialize())
		utils.Handle(err)
		err = transaction.Set([]byte("lh"), newBlock.Hash)
		// 记得更新BlockChain的LastHash字段
		bc.LastHash = newBlock.Hash
		return err
	})
	utils.Handle(err)
}

// 初始化我们的区块链并创建一个数据库保存
func InitBlockChain(address []byte) *BlockChain {
	var lastHash []byte

	// 先检查是否有存储区块链的数据库存在
	if utils.FileExists(constcoe.BCFile) {
		fmt.Println("blockchain already exists")
		runtime.Goexit()
	}

	// opts是启动badger的配置，地址指定为constcoe.BCPath
	opts := badger.DefaultOptions(constcoe.BCPath)
	// 使数据库的操作信息不输出到标准输出中
	opts.Logger = nil

	// 按照配置启动数据库并返回该数据库的指针
	db, err := badger.Open(opts)
	utils.Handle(err)

	// 对数据库进行更新操作
	// db.Updata构造了一个事务给内部函数调用，然后在外部实现事务的提交与结束
	err = db.Update(func(txn *badger.Txn) error {
		// 创建创世区块
		genesis := GenesisBlock(address)
		fmt.Println("Genesis Created")
		// 向数据库中添加(创世区块哈希，创世区块序列化字节串)键值对
		err = txn.Set(genesis.Hash, genesis.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash) //store the hash of the block in blockchain
		utils.Handle(err)
		err = txn.Set([]byte("ogprevhash"), genesis.PrevHash) //store the prevhash of genesis(original) block
		utils.Handle(err)
		lastHash = genesis.Hash
		return err
	})
	utils.Handle(err)
	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

// 读取已有的数据库并加载区块链
func ContinueBlockChain() *BlockChain {
	// 查看是否有存储区块链的数据库存在
	if utils.FileExists(constcoe.BCFile) == false {
		fmt.Println("No blockchain found, please create one first")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(constcoe.BCPath)
	opts.Logger = nil
	db, err := badger.Open(opts)
	utils.Handle(err)

	// 使用db.View()函数来调取视图，读取当前区块链的最后一个区块的哈希值
	err = db.View(func(txn *badger.Txn) error {
		// 查询键"lh"
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)
		return err
	})
	utils.Handle(err)

	chain := BlockChain{lastHash, db}
	return &chain
}

// 根据目标地址寻找可用交易信息
func (bc *BlockChain) FindUnspentTransactions(address []byte) []transaction.Transaction {
	// 要返回的包含指定地址的可用交易信息的切片
	var unSpentTxs []transaction.Transaction

	// spentTxs用于记录遍历区块链时那些已经被使用的交易信息的Output
	// key值为交易信息的ID值（需要转成string），value值为Output在该交易信息中的序号，注意这是个[]int类型
	spentTxs := make(map[string][]int)

	// 该循环内部: 向后遍历区块——>遍历某区块的交易——>遍历某交易的Output——>遍历spentTxs对应键为txID的值
	// 从后向前遍历区块链非常重要，这样使得所有交易按从晚到早的顺序被检索，不会出现遗漏
	for idx := len(bc.Blocks) - 1; idx >= 0; idx-- {
		// 从最后一个区块开始向前遍历区块链，然后遍历每一个区块中的交易信息
		block := bc.Blocks[idx]
		// tx是某个区块中的某个交易
		for _, tx := range block.Transactions {
			// txID是其中一个交易的ID值
			txID := hex.EncodeToString(tx.ID)

			// Go语言Label特性
		IterOutputs:
			// 遍历某个交易信息的Output，如果该Output在spentTxs中就跳过，说明该Output已被消费
			// outIdx表示该Output在某个交易中的序号
			// out表示该Output的地址(指针)
			for outIdx, out := range tx.Outputs {
				// 若spentTxs[txID]不为空则说明这个交易有些Output已经被消耗
				if spentTxs[txID] != nil {
					// 遍历spentTxs对应键为txID的值(值的类型为[]int)
					for _, spentOut := range spentTxs[txID] {
						// 若spentOut和outIdx一致，说明该Output已经被记录/消费，则开始下一个IterOutputs循环
						if spentOut == outIdx {
							continue IterOutputs
						}
					}
				}

				// 若该Output没有被消费则确认ToAddress正确与否，正确就是我们要找的可用交易信息
				if out.ToAddressRight(address) {
					unSpentTxs = append(unSpentTxs, *tx)
				}
			}

			// 将该交易的所有Output录入到unSpentTxs之后，检索该交易是否输入有address，即address是否有使用过某个output
			// 检查当前交易信息是否为Base Transaction
			// 如果不是就检查当前交易信息的input中是否包含目标地址，有的话就将指向的Output信息加入到spentTxs中
			// 还需要多理解
			if !tx.IsBase() {
				for _, in := range tx.Inputs {
					if in.FromAddressRight(address) {
						inTxID := hex.EncodeToString(in.TxID)
						// 该交易对应的Input中标记的交易ID作为键计入spentTxs，键对应的值为Input中的OutIdx
						spentTxs[inTxID] = append(spentTxs[inTxID], in.OutIdx)
					}
				}
			}
		}
	}
	return unSpentTxs
}

// 找到一个地址的完成交易需要的UTXO
func (bc *BlockChain) FindSpendableOutputs(address []byte, amount int) (int, map[string]int) {
	// 键为交易ID，值为该交易第几个输出
	// 一个交易中最多只有一个Output指向Address
	unspentOuts := make(map[string]int)
	// 地址address所以未花费的输出
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	// 遍历address的UTXO
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		// 遍历每个交易的Output
		for outIdx, out := range tx.Outputs {
			// 若Output的地址为address且当前收集到的资产总和小于amount则将此时out.Value加到accumulated中
			if out.ToAddressRight(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = outIdx
				if accumulated >= amount {
					break Work
				}
				continue Work // one transaction can only have one output referred to adderss
			}
		}
	}
	return accumulated, unspentOuts
}

// 构建交易
func (bc *BlockChain) CreateTransaction(from, to []byte, amount int) (*transaction.Transaction, bool) {
	var inputs []transaction.TxInput
	var outputs []transaction.TxOutput

	// acc为累计的UTXO总和，validOutputs为UTXO列表
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	// UTXO总和不够则交易失败
	if acc < amount {
		fmt.Println("Not enough coins!")
		return &transaction.Transaction{}, false
	}
	for txid, outidx := range validOutputs {
		txID, err := hex.DecodeString(txid)
		utils.Handle(err)
		input := transaction.TxInput{txID, outidx, from}
		inputs = append(inputs, input)
	}

	// 目前Output的转账人最多只有两个(一个别人一个自己)
	outputs = append(outputs, transaction.TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, transaction.TxOutput{acc - amount, from})
	}
	tx := transaction.Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx, true
}

//
func (bc *BlockChain) Mine(txs []*transaction.Transaction) {
	bc.AddBlock(txs)
}
