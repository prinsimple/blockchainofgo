package pos

import (
	"crypto/sha256"
	"math/big"
	"time"
)

type Validator struct {
	Address []byte
	Stake   uint64
}

// 验证者池
type ValidatorPool struct {
	Validators []*Validator
	TotalStake uint64
}

// 添加验证者
func (pool *ValidatorPool) AddValidator(validator *Validator) {
	pool.Validators = append(pool.Validators, validator)
	pool.TotalStake += validator.Stake
}

// 选择验证者
func (pool *ValidatorPool) SelectValidator(prevHash []byte) *Validator {
	// 使用前一个区块的哈希作为随机种子
	seed := time.Now().Unix()
	data := append(prevHash, []byte(string(seed))...)
	hash := sha256.Sum256(data)

	// 将哈希转换为大整数
	hashInt := new(big.Int).SetBytes(hash[:])

	// 根据总权益创建范围
	totalStakeInt := new(big.Int).SetUint64(pool.TotalStake)

	// 选择点 = 哈希值 % 总权益
	position := new(big.Int).Mod(hashInt, totalStakeInt)

	// 遍历验证者池找到选中的验证者
	current := new(big.Int)
	for _, validator := range pool.Validators {
		current.Add(current, new(big.Int).SetUint64(validator.Stake))
		if current.Cmp(position) > 0 {
			return validator
		}
	}

	// 默认返回第一个验证者（不应该发生）
	return pool.Validators[0]
}
