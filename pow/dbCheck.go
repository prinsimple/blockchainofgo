package pow

import (
	"fmt"

	"github.com/dgraph-io/badger"
	// "log"
)

func DbCheck() {
	// 打开数据库连接
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	Handle(err)
	defer db.Close() // 确保函数结束时关闭数据库

	// 创建只读事务查看数据库内容
	err = db.View(func(txn *badger.Txn) error {
		// 设置迭代器选项
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close() // 确保迭代器在函数结束时关闭

		// 遍历数据库中的所有键值对
		for it.Seek([]byte{}); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			// 获取值的副本
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			// 打印键值对
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
	Handle(err)
}
