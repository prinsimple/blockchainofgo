package pos

import (
	"bytes"
	"crypto/sha256"
	"time"
)

// 区块结构体
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Timestamp    int64
	Validator    []byte // 验证者的地址
	Stake        uint64 // 验证者的权益数量
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
func CreateBlock(txs []*Transaction, prevHash []byte, validator []byte, stake uint64) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     prevHash,
		Timestamp:    time.Now().Unix(),
		Validator:    validator,
		Stake:        stake,
	}

	block.Hash = block.calculateHash()
	return block
}

// 计算区块哈希
func (b *Block) calculateHash() []byte {
	timestamp := []byte(string(b.Timestamp))
	headers := bytes.Join(
		[][]byte{
			b.PrevHash,
			b.TransactionsHash(),
			timestamp,
			b.Validator,
		},
		[]byte{},
	)
	hash := sha256.Sum256(headers)
	return hash[:]
}

// 创建创世区块
func CreateGenesisBlock(tx *Transaction, validator []byte, stake uint64) *Block {
	return CreateBlock([]*Transaction{tx}, []byte{}, validator, stake)
}
