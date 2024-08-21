package server

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"errors"
)

var NoCorrectPeerAvailable = errors.New("There are no correct peers available")

type PeerSet struct {
	peer []*P2PClient
}

func (p *PeerSet) AddPeer(IPAddress string, portNumber int, connectionCount int) error {

	client, err := NewClient(IPAddress, portNumber, connectionCount)
	if err != nil {
		return err
	}

	// starts the main loop of client
	go client.Start()

	p.peer = append(p.peer, client)

	return nil
}

func (p *PeerSet) DissaminateChunks(chunks []internal.Chunk) {
	for index, chunk := range chunks {
		peer := p.selectPeer(index)
		peer.SendBlockChunk(chunk)
	}
}

func (p *PeerSet) ForwardChunk(chunk internal.Chunk) {

	forwardCount := 0
	for _, peer := range p.peer {
		if peer.err != nil {
			continue
		}
		forwardCount++
		peer.SendBlockChunk(chunk)
	}

	if forwardCount == 0 {
		panic(NoCorrectPeerAvailable)
	}
}

func (p *PeerSet) selectPeer(index int) *P2PClient {

	peerCount := len(p.peer)
	for i := 0; i < peerCount; i++ {
		peer := p.peer[(index+i)%peerCount]
		if peer.err == nil {
			return peer
		}
	}

	panic(NoCorrectPeerAvailable)
}
