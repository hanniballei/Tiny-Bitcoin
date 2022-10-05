package utils

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

// 简单的错误处理函数
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// 该函数将int64转为字节串类型
func ToHexInt(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// 检查文件地址下文件是否存在的函数
func FileExists(fileAddr string) bool {
	if _, err := os.Stat(fileAddr); os.IsNotExist(err) {
		return false
	}
	return true
}
