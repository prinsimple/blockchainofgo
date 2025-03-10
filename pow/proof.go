package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"
	"sync"
)

var (
	targetBits      = 24 // 设置默认挖矿难度
	difficultyMutex sync.RWMutex
)

// SetDifficulty 设置新的挖矿难度
func SetDifficulty(newDifficulty int) {
	difficultyMutex.Lock()
	defer difficultyMutex.Unlock()
	targetBits = newDifficulty
}

// GetDifficulty 获取当前挖矿难度
func GetDifficulty() int {
	difficultyMutex.RLock()
	defer difficultyMutex.RUnlock()
	return targetBits
}

// 工作量证明结构体
type ProofOfWork struct {
	block  *Block   // 区块
	target *big.Int // 目标值
}

// 创建工作量证明
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{block: b, target: target}
}

// 准备数据
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevHash,
			pow.block.TransactionsHash(),
			IntToHex(int64(nonce)),
			IntToHex(int64(targetBits)),
		},
		[]byte{},
	)
	return data
}

// 运行工作量证明
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

// 验证工作量证明
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}

// ToHex 将整数转换为十六进制字节数组
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
