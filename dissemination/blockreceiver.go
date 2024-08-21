package dissemination

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"fmt"
	"sort"
)

type BlockReceiver struct {
	blockCount         int
	chunkCount         int
	dataChunkCount     int
	blockMap           map[string][]internal.Chunk
	recievedChunkCount int
}

func NewBlock(leaderCount int, chunkCount int, dataChunkCount int) *BlockReceiver {
	r := &BlockReceiver{
		blockCount:     leaderCount,
		chunkCount:     chunkCount,
		dataChunkCount: dataChunkCount,
		blockMap:       make(map[string][]internal.Chunk),
	}
	return r
}

// AddChunk stores a chunk of block to recostruct the whole block later
func (r *BlockReceiver) AddChunk(chunk internal.Chunk) {
	key := string(chunk.Issuer)
	chunkSlice := r.blockMap[key]
	r.blockMap[key] = append(chunkSlice, chunk)
}

// ReceivedAll checks whether all chunks are recived or not to reconstruct the blocks of a round
func (r *BlockReceiver) ReceivedAll() bool {

	if len(r.blockMap) != r.blockCount {
		return false
	}

	for _, chunkSlice := range r.blockMap {
		if len(chunkSlice) != r.dataChunkCount {
			return false
		}
	}
	return true
}

// GetBlocks recunstruct blocks using chunks, and returns the blocks by sorting the resulting block slice according to block hashes
func (r *BlockReceiver) GetBlocks() []internal.Message {

	if r.ReceivedAll() == false {
		panic(fmt.Errorf("not received all block chunks to reconstruct block/s"))
	}

	keys := make([]string, 0, len(r.blockMap))
	for k := range r.blockMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var messages []internal.Message

	for _, key := range keys {

		receivedChunks := r.blockMap[key]
		sort.Slice(receivedChunks, func(i, j int) bool {
			return receivedChunks[i].ChunkIndex < receivedChunks[j].ChunkIndex
		})

		block := internal.MergeChunks(receivedChunks, r.chunkCount, r.dataChunkCount)
		messages = append(messages, block)
	}

	return messages
}
