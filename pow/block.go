package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"os"
)

// 区块结构体
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

// 计算区块中所有交易的hash
func (b *Block) TransactionsHash() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// 创建区块
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 创建创世区块
func CreateGenesisBlock(tx *Transaction) *Block {
	return CreateBlock([]*Transaction{tx}, []byte{})
}

// Handle 处理错误
// 如果有错误则触发 panic
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// HandleError 处理错误并输出发生错误的函数名和目录
// err: 要处理的错误
// funcName: 发生错误的函数名
func HandleError(err error, funcName string) {
	if err != nil {
		dir, errDir := os.Getwd()
		if errDir != nil {
			log.Panicf("Error getting directory: %s", errDir)
		} else {
			log.Panicf("Error occurred in function %s at: %s. Error: %s", funcName, dir, err)
		}
	}
}

// Serialize 将区块序列化为字节数组
// 使用 gob 编码器进行序列化
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	Handle(err)
	return res.Bytes()
}

// DeSerialize 将字节数组反序列化为区块
// data: 要反序列化的字节数组
// 返回反序列化后的区块
func (b *Block) DeSerialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)
	return &block
}
