package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Node struct {
	id       int
	peers    []*Node
	received map[string]bool
	mu       sync.Mutex
}

func NewNode(id int) *Node {
	return &Node{
		id:       id,
		peers:    []*Node{},
		received: make(map[string]bool),
	}
}

func (n *Node) AddPeer(peer *Node) {
	n.peers = append(n.peers, peer)
}

func (n *Node) Gossip(transaction string) {
	n.mu.Lock()
	if n.received[transaction] {
		n.mu.Unlock()
		return
	}
	n.received[transaction] = true
	n.mu.Unlock()

	fmt.Printf("Node %d received transaction: %s\n", n.id, transaction)

	// Simulate random gossip to a subset of peers
	for _, peer := range n.peers {
		go func(peer *Node) {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			peer.Gossip(transaction)
		}(peer)
	}
}

func main() {
	// Create nodes
	nodeCount := 5
	nodes := make([]*Node, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodes[i] = NewNode(i)
	}

	// Connect nodes as peers (fully connected for simplicity)
	for i := 0; i < nodeCount; i++ {
		for j := 0; j < nodeCount; j++ {
			if i != j {
				nodes[i].AddPeer(nodes[j])
			}
		}
	}

	// Start a transaction and let it spread via gossip
	transaction := "tx12345"
	fmt.Println("Starting transaction propagation...")
	nodes[0].Gossip(transaction)

	// Allow time for gossip to spread
	time.Sleep(2 * time.Second)
	fmt.Println("Transaction propagation completed.")
}
