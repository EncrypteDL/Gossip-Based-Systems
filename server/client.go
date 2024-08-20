package server

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"fmt"
	"net/rpc"
)

// Client implement P2P client
type P2PClient struct {
	IPAddress  string
	PortNumber int

	rpcClient  []*rpc.Client
	blockChunk chan internal.Chunk

	err error
}

// NewClient creates a new client
func NewClient(IPAddress string, PortNumber int, connectionCount int) (*P2PClient, error) {
	if connectionCount < 1 {
		panic(fmt.Errorf("connection count is %d, it must be bigger than 1", connectionCount))
	}

	var clients []*rpc.Client
	for i := 0; i < connectionCount; i++ {
		rpcClient, err := rpc.Dial("TCP", fmt.Sprintf("%s:%d", IPAddress, PortNumber))
		if err != nil {
			return nil, err
		}
		clients = append(clients, rpcClient)
	}

	client := &P2PClient{}
	client.IPAddress = IPAddress
	client.PortNumber = PortNumber
	client.rpcClient = append(client.rpcClient, clients...)

	client.blockChunk = make(chan internal.Chunk, 1012)

	return nil, client.err
}

// Start starts main loop of client. It Blocks teh calling goroutine
func (c *P2PClient) Start() {
	c.mainLoop()
}

// SendBlockChunk enqueues a chunk of a block to send
func (c *P2PClient) SendBlockChunk(chunk internal.Chunk) {
	c.blockChunk <- chunk
}

func (c *P2PClient) mainLoop() {
	var SendChunkCount int64
	connectionCount := int64(len(c.rpcClient))

	for {
		SendChunkCount++
		rpcClient := c.rpcClient[SendChunkCount%connectionCount]
		blockChunk := <-c.blockChunk

		go func() {
			err := rpcClient.Call("P2P Server.HandleblockChunk", blockChunk, nil)
			//TODO: needs to handle the error properly
			if err != nil {
				panic(err)
			}
		}()
	}
}
