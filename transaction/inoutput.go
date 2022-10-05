package transaction

import "bytes"

// 交易信息的output
type TxOutput struct {
	Value     int    // 转出资产值
	ToAddress []byte // 资产接收者的地址
}

// 交易信息的input
type TxInput struct {
	TxID        []byte // 支持本次交易的前置交易信息
	OutIdx 	    int    // 前置交易信息中的第几个Output
	FromAddress []byte // 资产转出者的地址
}

// 验证FromAddress是否正确
func (in *TxInput) FromAddressRight (address []byte) bool {
	return bytes.Equal(in.FromAddress, address)
}

// 验证ToAddress是否正确
func (out *TxOutput) ToAddressRight(address []byte) bool {
	return bytes.Equal(out.ToAddress, address)
}