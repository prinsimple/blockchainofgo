// network 包提供了网络传输相关的接口和实现
package network

import (
	"fmt"
	"sync"
)

// LocalTransport 实现了本地传输层的功能，用于模拟网络通信
type LocalTransport struct {
	// addr 表示本地传输层的地址，用于标识不同的节点
	addr NetAddr
	// consumeCh 是一个通道，用于接收RPC消息
	// 当其他节点发送消息时，消息会通过这个通道传递
	consumeCh chan RPC
	// lock 是一个读写锁，用于保护peers map的并发访问
	// 使用读写锁而不是普通互斥锁，因为读操作（发送消息）比写操作（建立连接）更频繁
	lock sync.RWMutex
	// peers 存储所有已连接的本地传输层节点
	// key是节点的地址，value是对应的LocalTransport实例
	peers map[NetAddr]*LocalTransport
}

// NewLocalTransport 创建一个新的本地传输层实例
// 参数addr指定了该传输层的地址
func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC),                    // 创建一个新的通道用于接收消息
		peers:     make(map[NetAddr]*LocalTransport), // 初始化peers map
	}
}

// Consume 返回用于接收RPC消息的通道
// 其他节点可以通过这个通道发送消息给当前节点
func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

// Connect 用于建立与其他本地传输层的连接
// 参数tr是要连接的另一个LocalTransport实例
func (t *LocalTransport) Connect(tr Transport) error {
	// 使用写锁保护peers map的修改操作
	trans := tr.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()

	// 将目标传输层添加到peers map中
	t.peers[tr.Addr()] = trans
	return nil
}

// Addr 返回本地传输层的地址
// 用于标识当前节点
func (t *LocalTransport) Addr() NetAddr {
	return t.addr
}

// SendMessage 向指定地址发送消息
// 参数to是目标节点的地址，payload是要发送的消息内容
func (t *LocalTransport) SendMessage(to NetAddr, payload []byte) error {
	// 使用读锁保护peers map的访问
	// 使用读锁而不是写锁，因为只是读取操作
	t.lock.RLock()
	defer t.lock.RUnlock()

	// 查找目标地址对应的peer
	peer, ok := t.peers[to]
	if !ok {
		// 如果找不到peer，返回错误
		return fmt.Errorf("%s : Could not send message to %s", t.addr, to)
	}

	// 向peer的consumeCh通道发送RPC消息
	// 消息包含发送方地址和消息内容
	peer.consumeCh <- RPC{
		From:    t.addr,  // 设置发送方地址
		Payload: payload, // 设置消息内容
	}

	return nil
}
