// network 包提供了网络传输相关的接口和实现
package network

// NetAddr 表示网络地址的类型
type NetAddr string

// RPC 结构体定义了远程过程调用的基本结构
type RPC struct {
	// From 表示消息的发送方地址
	From NetAddr
	// Payload 包含实际的消息内容
	Payload []byte
}

// Transport 接口定义了网络传输层需要实现的基本方法
type Transport interface {
	// Consume 返回一个只读通道，用于接收RPC消息
	Consume() <-chan RPC
	// Connect 用于建立与其他传输层的连接
	Connect(Transport) error
	// SendMessage 向指定地址发送消息
	SendMessage(NetAddr, []byte) error
	// Addr 返回当前传输层的地址
	Addr() NetAddr
}
