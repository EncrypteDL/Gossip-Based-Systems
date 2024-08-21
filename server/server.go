package server

import "EncrypteDL/Gossip-Based-Systems/internal"

type P2PServer struct{
	demux *internal.Demux
}

func NewServer(demux *internal.Demux) *P2PServer {
	server := &P2PServer{demux: demux}
	return server
}

func (s *P2PServer) HandleBlockChunk(chunk *internal.Chunk, reply *int) error {

	s.demux.EnqueBlockChunk(*chunk)

	return nil
}