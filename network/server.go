// network 包提供了网络传输相关的接口和实现
package network

import (
	"fmt"
)

// ServerOpts 定义了服务器配置选项
type ServerOpts struct {
	// Transports 包含服务器使用的所有传输层实现
	// 一个服务器可以有多个传输层，比如同时支持TCP和UDP
	Transports []Transport
}

// Server 表示网络服务器
type Server struct {
	// ServerOpts 包含服务器的配置选项
	ServerOpts
	// peers 存储所有已连接的节点
	// key是节点的地址，value是节点的信息
	peers map[NetAddr]*Peer
}

// Peer 表示一个已连接的节点
type Peer struct {
	// Transport 是节点使用的传输层实现
	Transport
	// Incoming 是一个通道，用于接收来自该节点的消息
	Incoming chan RPC
}

// NewServer 创建一个新的服务器实例
// 参数opts指定了服务器的配置选项
func NewServer(opts ServerOpts) *Server {
	return &Server{
		ServerOpts: opts,
		peers:      make(map[NetAddr]*Peer),
	}
}

// Start 启动服务器
// 开始监听所有传输层的消息
func (s *Server) Start() {
	// 遍历所有传输层
	for _, transport := range s.Transports {
		// 启动一个goroutine来处理每个传输层的消息
		go s.handleTransport(transport)
	}
}

// handleTransport 处理指定传输层的消息
// 参数transport是要处理的传输层
func (s *Server) handleTransport(transport Transport) {
	// 获取传输层的消息通道
	rpcCh := transport.Consume()
	// 持续接收消息
	for rpc := range rpcCh {
		// 处理接收到的消息
		s.handleRPC(rpc)
	}
}

// handleRPC 处理接收到的RPC消息
// 参数rpc是要处理的RPC消息
func (s *Server) handleRPC(rpc RPC) {
	// 获取发送方地址
	from := rpc.From
	// 查找或创建发送方节点
	peer, ok := s.peers[from]
	if !ok {
		// 如果节点不存在，创建新节点
		peer = &Peer{
			Transport: nil, // TODO: 需要实现获取发送方传输层的方法
			Incoming:  make(chan RPC),
		}
		s.peers[from] = peer
	}
	// 将消息发送到节点的Incoming通道
	peer.Incoming <- rpc
}

// Connect 连接到指定地址的节点
// 参数addr是要连接的节点地址
func (s *Server) Connect(addr NetAddr) error {
	// 遍历所有传输层
	for _, transport := range s.Transports {
		// 尝试使用每个传输层建立连接
		if err := transport.Connect(transport); err != nil {
			return fmt.Errorf("failed to connect to %s: %v", addr, err)
		}
	}
	return nil
}

// SendMessage 向指定地址发送消息
// 参数to是目标节点地址，payload是要发送的消息内容
func (s *Server) SendMessage(to NetAddr, payload []byte) error {
	// 查找目标节点
	peer, ok := s.peers[to]
	if !ok {
		return fmt.Errorf("peer %s not found", to)
	}
	// 使用节点的传输层发送消息
	return peer.SendMessage(to, payload)
}
