package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"goblockchain/constcoe"
	"goblockchain/utils"
)

type Transaction struct {
	ID      []byte     // 自身哈希值
	Inputs  []TxInput  // 用于标记支持我们本次转账的前置的交易信息的TxOutput
	Outputs []TxOutput // 记录我们本次转账的amount和Reciever
}

// 计算每个transaction的哈希值
func (tx *Transaction) TxHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	return hash[:]
}

func (tx *Transaction) SetID() {
	tx.ID = tx.TxHash()
}

// 类似初始区块一样，我们需要一个初始交易
func BaseTx(toaddress []byte) *Transaction {
	txIn := TxInput{[]byte{}, -1, []byte{}}
	txOut := TxOutput{constcoe.InitCoin, toaddress}
	tx := Transaction{[]byte("This is the Base Transaction!"), []TxInput{txIn}, []TxOutput{txOut}}
	return &tx
}

// 判断某个交易是否为初始交易
func (tx *Transaction) IsBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].OutIdx == -1
}
