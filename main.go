package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"blockchainofgo/pow"
)

// CommandLine 命令行结构体
type CommandLine struct{}

func main() {
	cli := &CommandLine{}
	cli.run()
}

func (cli *CommandLine) printUsage() {
	fmt.Println("欢迎使用区块链!")
	fmt.Println("使用说明:")
	fmt.Println(" init -address 地址 - 创建一个新的区块链并发送创世区块奖励到指定地址")
	fmt.Println(" balance -address 地址 - 获取指定地址的余额")
	fmt.Println(" send -from 发送地址 -to 接收地址 -amount 金额 - 发送金额从一个地址到另一个地址")
	fmt.Println(" getchain - 打印区块链中的所有区块")
	fmt.Println(" mine - 挖掘新区块")
	fmt.Println(" getpending - 获取待处理的交易")
	fmt.Println(" difficulty - 获取当前挖矿难度")
	fmt.Println(" setdifficulty -value 难度值 - 设置新的挖矿难度")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("init", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("getchain", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	getPendingCmd := flag.NewFlagSet("getpending", flag.ExitOnError)
	getDifficultyCmd := flag.NewFlagSet("difficulty", flag.ExitOnError)
	setDifficultyCmd := flag.NewFlagSet("setdifficulty", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "要查询余额的地址")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "接收创世区块奖励的地址")
	sendFrom := sendCmd.String("from", "", "发送方地址")
	sendTo := sendCmd.String("to", "", "接收方地址")
	sendAmount := sendCmd.Int("amount", 0, "发送金额")
	setDifficultyValue := setDifficultyCmd.Int("value", 0, "新的难度值")

	switch os.Args[1] {
	case "balance":
		err := getBalanceCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "init":
		err := createBlockchainCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "getchain":
		err := printChainCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "mine":
		err := mineCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "getpending":
		err := getPendingCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "difficulty":
		err := getDifficultyCmd.Parse(os.Args[2:])
		pow.Handle(err)
	case "setdifficulty":
		err := setDifficultyCmd.Parse(os.Args[2:])
		pow.Handle(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if mineCmd.Parsed() {
		cli.mine()
	}

	if getPendingCmd.Parsed() {
		cli.getPendingTransactions()
	}

	if getDifficultyCmd.Parsed() {
		cli.getDifficulty()
	}

	if setDifficultyCmd.Parsed() {
		if *setDifficultyValue <= 0 {
			setDifficultyCmd.Usage()
			runtime.Goexit()
		}
		cli.setDifficulty(*setDifficultyValue)
	}
}

func (cli *CommandLine) getDifficulty() {
	difficulty := pow.GetDifficulty()
	fmt.Printf("%d", difficulty)
}

func (cli *CommandLine) setDifficulty(value int) {
	pow.SetDifficulty(value)
	fmt.Printf("难度已更新为: %d", value)
}

func initBlockchain() *pow.BlockChain {
	if pow.DbExists() {
		fmt.Println("区块链已存在")
		return pow.ContinueBlockChain("")
	}
	return nil
}

func createGenesisBlock(bc *pow.BlockChain, address string) {
	if pow.DbExists() {
		fmt.Println("区块链已存在")
		return
	}

	bc = pow.InitBlockChain(address)
	if bc == nil {
		fmt.Println("创建创世区块失败")
		return
	}
	defer bc.DataBase.Close()
	fmt.Println("创世区块已创建")
}

func getBlockchain(bc *pow.BlockChain) {
	if bc == nil {
		fmt.Println("区块链未初始化")
		return
	}
	iter := bc.Iterator()
	var blocks []*pow.Block
	for {
		block := iter.Next()
		blocks = append(blocks, block)
		if len(block.PrevHash) == 0 {
			break
		}
	}
	blocksJSON, _ := json.Marshal(blocks)
	fmt.Println(string(blocksJSON))
}

func getBalance(bc *pow.BlockChain, address string) int {
	balance := 0
	UTXOs := bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}

func (cli *CommandLine) getBalance(address string) {
	bc := pow.ContinueBlockChain("")
	if bc == nil {
		return
	}
	defer bc.DataBase.Close()
	balance := getBalance(bc, address)
	fmt.Printf("地址 %s 的余额为: %d\n", address, balance)
}

func (cli *CommandLine) createBlockChain(address string) {
	if pow.DbExists() {
		fmt.Println("区块链已存在")
		return
	}

	bc := pow.InitBlockChain(address)
	if bc == nil {
		fmt.Println("创建创世区块失败")
		return
	}
	defer bc.DataBase.Close()
	fmt.Println("创世区块已创建")
}

func (cli *CommandLine) send(from, to string, amount int) {
	bc := pow.ContinueBlockChain("")
	if bc == nil {
		return
	}
	defer bc.DataBase.Close()
	tx := pow.NewTransaction(from, to, amount, bc)
	if tx == nil {
		fmt.Println("创建交易失败")
		return
	}
	bc.AddToPool(tx)
	fmt.Println("交易已添加到待处理池")
}

func (cli *CommandLine) printChain() {
	bc := pow.ContinueBlockChain("")
	if bc == nil {
		return
	}
	defer bc.DataBase.Close()
	getBlockchain(bc)
}

func (cli *CommandLine) mine() {
	bc := pow.ContinueBlockChain("")
	if bc == nil {
		return
	}
	defer bc.DataBase.Close()

	// 显示挖矿动画
	fmt.Print("\n开始挖矿")
	done := make(chan bool)
	go func() {
		frames := []string{"⛏️ ", "⛏️  ", "⛏️   ", "⛏️    ", "⛏️     "}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r正在挖矿中 %s", frames[i])
				time.Sleep(200 * time.Millisecond)
				i = (i + 1) % len(frames)
			}
		}
	}()

	// 创建挖矿奖励交易
	minerAddress := "miner" // 这里可以改成从参数传入
	cbtx := pow.CoinBaseTx(minerAddress, "")
	// 获取待处理交易
	txs := bc.GetPendingTransactions()
	// 将挖矿奖励交易加入到待打包的交易中
	txs = append([]*pow.Transaction{cbtx}, txs...)
	// 创建新区块
	bc.AddBlock(txs)

	// 停止动画
	done <- true
	fmt.Printf("\r✨ 新区块已挖出！\n")
}

func (cli *CommandLine) getPendingTransactions() {
	bc := pow.ContinueBlockChain("")
	if bc == nil {
		return
	}
	defer bc.DataBase.Close()
	txs := bc.GetPendingTransactions()
	txsJSON, err := json.Marshal(txs)
	if err != nil {
		fmt.Println("序列化交易失败")
		return
	}
	fmt.Println(string(txsJSON))
}
