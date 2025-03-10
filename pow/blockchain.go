// Package blockchain implements a simple blockchain and utilities.
//
// BlockChain provides methods to initialize a new blockchain, add blocks,
// iterate over the chain, and find spendable outputs.
//
// BlockChainIterator allows iterating over the blockchain.
//
// The InitBlockChain, AddBlock, FindSpendableOutputs and other functions
// implement the core blockchain logic.
package pow

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"log"

	// "log"
	// "errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "db/blocks"
	dbFile      = "db/blocks/MANIFEST"
	genesisData = "First transaction from Genesis"
	pendingKey  = "pending" // 待处理交易的键
)

// BlockChain 区块链结构体
// LastHash 存储最后一个区块的哈希值
// DataBase 存储区块数据的数据库实例
// PendingTransactions 待处理的交易池
type BlockChain struct {
	LastHash            []byte
	DataBase            *badger.DB
	PendingTransactions []*Transaction
}

// BlockChainIterator 区块链迭代器
// CurrentHash 当前遍历到的区块哈希值
// DataBase 区块链数据库实例
type BlockChainIterator struct {
	CurrentHash []byte
	DataBase    *badger.DB
}

// DbExists 检查区块链数据库是否已存在
func DbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// InitBlockChain 初始化一个新的区块链
// 创建创世区块和区块链数据库
// address: 接收创世区块奖励的地址
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DbExists() {
		fmt.Println("DB already exists")
		runtime.Goexit()
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = os.MkdirAll(dbPath, 0755)
		HandleError(err, "InitBlockChain")
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	HandleError(err, "InitBlockChain")
	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinBaseTx(address, genesisData)
		genesis := CreateGenesisBlock(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		HandleError(err, "InitBlockChain")
		err = txn.Set([]byte("lh"), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})

	HandleError(err, "InitBlockChain")

	blockchain := BlockChain{
		LastHash:            lastHash,
		DataBase:            db,
		PendingTransactions: make([]*Transaction, 0),
	}
	return &blockchain
}

// AddBlock 向区块链中添加新的区块
// transactions: 要打包进区块的交易列表
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	// 如果没有提供交易，使用待处理池中的交易
	if len(transactions) == 0 {
		transactions = chain.GetPendingTransactions()
	}

	var lastHash []byte
	err := chain.DataBase.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleError(err, "AddBlock")
		lastHash, err = item.ValueCopy(nil)
		return err
	})
	HandleError(err, "AddBlock")

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.DataBase.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleError(err, "AddBlock")
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	HandleError(err, "AddBlock")

	// 清空待处理池
	chain.ClearPool()
}

// ContinueBlockChain 加载已存在的区块链
// 如果区块链不存在则退出程序
// address: 挖矿奖励接收地址
func ContinueBlockChain(address string) *BlockChain {
	if !DbExists() {
		fmt.Println("区块链不存在，请先创建")
		return nil
	}

	var lastHash []byte
	var db *badger.DB
	var err error

	// 添加重试机制
	for i := 0; i < 3; i++ {
		opts := badger.DefaultOptions(dbPath)
		opts.Logger = nil
		db, err = badger.Open(opts)
		if err == nil {
			break
		}
		log.Printf("尝试打开数据库失败 (尝试 %d/3): %v", i+1, err)
		time.Sleep(time.Second) // 等待1秒后重试
	}

	if err != nil {
		log.Printf("无法打开数据库: %v", err)
		return nil
	}

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	if err != nil {
		db.Close()
		log.Printf("获取最后区块哈希失败: %v", err)
		return nil
	}

	chain := BlockChain{
		LastHash:            lastHash,
		DataBase:            db,
		PendingTransactions: make([]*Transaction, 0),
	}

	return &chain
}

// Iterator 创建一个区块链迭代器
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.DataBase}

	return iter
}

// Next 获取迭代器中的下一个区块
// 返回当前哈希对应的区块，并将迭代器移动到前一个区块
func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.DataBase.View(func(txn *badger.Txn) error {
		// if len(iter.CurrentHash) == 0 {
		// 	return errors.New("Error: Current Hash is empty")
		// }
		item, err := txn.Get(iter.CurrentHash)
		HandleError(err, "Next")
		encodedBlock, err := item.ValueCopy(nil)
		block = block.DeSerialize(encodedBlock)

		return err
	})
	HandleError(err, "Next")

	iter.CurrentHash = block.PrevHash

	return block
}

// FindUnspendTransactions 查找地址相关的所有未花费交易
// address: 要查询的地址
// 返回包含未花费输出的交易列表
func (chain *BlockChain) FindUnspendTransactions(address string) []Transaction {
	var unspendTxs []Transaction
	spentTXOs := make(map[string][]int)
	iter := chain.Iterator()
	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlock(address) {
					unspendTxs = append(unspendTxs, *tx)
				}
			}
			if !tx.isCoinBase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}

	}
	return unspendTxs
}

// FindUTXO 查找地址的所有未花费交易输出
// address: 要查询的地址
// 返回该地址能够使用的所有交易输出
func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspendTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlock(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// FindSpendableOutputs 查找地址中足够支付指定金额的未花费输出
// address: 要查询的地址
// amount: 需要支付的金额
// 返回累计金额和可用的交易输出映射
func (chain *BlockChain) FindSpendableOutputs(address string, ammount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspendTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlock(address) && accumulated < ammount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= ammount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

// AddToPool 添加交易到待处理池
func (chain *BlockChain) AddToPool(tx *Transaction) {
	// 先获取现有的待处理交易
	existingTxs := chain.GetPendingTransactions()

	// 将新交易添加到列表中
	allTxs := append(existingTxs, tx)

	// 将整个待处理池持久化到数据库
	err := chain.DataBase.Update(func(txn *badger.Txn) error {
		serializedTxs := SerializeTransactions(allTxs)
		if len(serializedTxs) == 0 {
			log.Printf("警告：序列化的交易数据为空")
			return nil
		}
		return txn.Set([]byte(pendingKey), serializedTxs)
	})

	if err != nil {
		log.Printf("添加交易到待处理池失败: %v", err)
	} else {
		log.Printf("成功添加交易到待处理池，当前待处理交易数量: %d", len(allTxs))
		// 更新内存中的待处理交易列表
		chain.PendingTransactions = allTxs
	}
}

// GetPendingTransactions 获取待处理的交易
func (chain *BlockChain) GetPendingTransactions() []*Transaction {
	var transactions []*Transaction

	// 从数据库中读取待处理交易
	err := chain.DataBase.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(pendingKey))
		if err == badger.ErrKeyNotFound {
			log.Printf("未找到待处理交易")
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			if len(val) > 0 {
				transactions = DeserializeTransactions(val)
				log.Printf("从数据库读取到 %d 笔待处理交易", len(transactions))
			} else {
				log.Printf("数据库中的待处理交易数据为空")
			}
			return nil
		})
	})

	if err != nil {
		log.Printf("获取待处理交易时发生错误: %v", err)
		return make([]*Transaction, 0)
	}

	// 更新内存中的待处理交易列表
	chain.PendingTransactions = transactions
	return transactions
}

// ClearPool 清空待处理交易池
func (chain *BlockChain) ClearPool() {
	chain.PendingTransactions = make([]*Transaction, 0)

	// 从数据库中删除待处理交易
	err := chain.DataBase.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(pendingKey))
	})
	HandleError(err, "ClearPool")
}

// SerializeTransactions 序列化交易列表
func SerializeTransactions(txs []*Transaction) []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// 注册Transaction类型
	gob.Register(&Transaction{})
	gob.Register(&TxInput{})
	gob.Register(&TxOutput{})

	// 直接序列化交易列表
	err := encoder.Encode(txs)
	if err != nil {
		log.Printf("序列化交易失败: %v", err)
		return []byte{}
	}
	return buf.Bytes()
}

// DeserializeTransactions 反序列化交易列表
func DeserializeTransactions(data []byte) []*Transaction {
	// 注册Transaction类型
	gob.Register(&Transaction{})
	gob.Register(&TxInput{})
	gob.Register(&TxOutput{})

	var transactions []*Transaction
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(&transactions)
	if err != nil {
		log.Printf("反序列化交易失败: %v", err)
		return make([]*Transaction, 0)
	}
	return transactions
}
