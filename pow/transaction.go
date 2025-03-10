package pow

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

// TxInput 表示交易输入，引用之前交易的输出
type TxInput struct {
	ID  []byte // 引用的交易ID
	Out int    // 引用的输出索引
	Sig string // 签名（简化版本，实际应使用数字签名）
}

// TxOutput 表示交易输出，包含金额和接收者
type TxOutput struct {
	Value  int    // 输出金额
	PubKey string // 接收者的公钥（简化版本，实际应使用公钥哈希）
}

// SetID 计算并设置交易的ID
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	// 使用gob编码交易数据
	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	Handle(err)

	// 计算交易数据的SHA-256哈希作为交易ID
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// NewTransaction 创建一个新的交易
func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	// 查找可用的未花费输出
	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("错误：余额不足")
	}

	// 创建交易输入
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// 创建交易输出（给接收者）
	outputs = append(outputs, TxOutput{amount, to})

	// 如果有找零，创建一个输出返回给发送者
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	// 创建交易并设置ID
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

// isCoinBase 判断交易是否为coinbase交易
func (tx *Transaction) isCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// CanUnlock 检查输入是否可以被指定数据解锁
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlock 检查输出是否可以被指定数据解锁
func (out *TxOutput) CanBeUnlock(data string) bool {
	return out.PubKey == data
}

// CoinBaseTx 创建一个coinbase交易（区块奖励交易）
func CoinBaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	// coinbase交易的输入没有引用之前的交易
	txInput := TxInput{[]byte{}, -1, data}
	// 创建一个输出，奖励50个币
	txOutput := TxOutput{50, to}

	tx := Transaction{nil, []TxInput{txInput}, []TxOutput{txOutput}}
	tx.SetID()
	return &tx
}
