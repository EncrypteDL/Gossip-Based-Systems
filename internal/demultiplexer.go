package internal

import (
	"fmt"
	"sync"
)

const (
	channelCapacity = 1024
)

// Demux provides message multiplexing service
// Network and consensus layer communicate using demux
type Demux struct {
	mutex sync.Mutex

	currentRound int

	// it is used to filter already processed messages
	processedMessageMap map[int]map[string]struct{}

	blockChunkChanMap map[int]chan Chunk
}

// NewDemultiplexer creates a new demultiplexer with initial round value
func NewDemultiplexer(initialRound int) *Demux {

	demux := &Demux{currentRound: initialRound}

	demux.processedMessageMap = make(map[int]map[string]struct{})
	demux.blockChunkChanMap = make(map[int]chan Chunk)

	return demux
}

// EnqueBlockChunk enques a block chunk to be the consumed by consensus layer
func (d *Demux) EnqueBlockChunk(chunk Chunk) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if chunk.Round < d.currentRound {
		// discarts a chunks because it belongs to a previous round
		return
	}

	chunkRound := chunk.Round
	chunkHash := string(chunk.Hash())
	if d.isProcessed(chunkRound, chunkHash) {
		// chunk is already processed
		return
	}

	chunkChan := d.getCorrespondingBlockChunkChan(chunkRound)
	chunkChan <- chunk

	// updates the queue length
	lengthChan := len(chunkChan)
	if lengthChan == 0 {
		lengthChan = 1
	}

	d.markAsProcessed(chunkRound, chunkHash)
}

// GetVoteBlockChunkChan returns Blockchunk channel
func (d *Demux) GetMessageChunkChan(round int) (chan Chunk, error) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if round < d.currentRound {
		return nil, fmt.Errorf("the current round value is bigger than the provided round value")
	}

	return d.getCorrespondingBlockChunkChan(round), nil
}

// UpdateRound updates the round.
// All messages blongs to the previous rounds discarted
// Update round mustbe called by an increased round number otherwise this function panics
func (d *Demux) UpdateRound(round int) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Round value should increase one by one
	if (round-d.currentRound) < 0 || (round-d.currentRound) > 1 {
		panic(fmt.Errorf("illegal round value, current round value %d, provided round value %d", d.currentRound, round))
	}

	d.currentRound = round
	d.deletePreviousRoundMessages()
}

// All the following functions are helper functions.
// They must be called from previous functions because
// they are not thread safe!

func (d *Demux) deletePreviousRoundMessages() {

	previousRound := d.currentRound - 1

	delete(d.processedMessageMap, previousRound)
	delete(d.blockChunkChanMap, previousRound)
}

func (d *Demux) getProcessedMessageMap(round int) map[string]struct{} {

	if val, ok := d.processedMessageMap[round]; ok {
		return val
	}

	val := make(map[string]struct{})
	d.processedMessageMap[round] = val

	return val
}

func (d *Demux) isProcessed(round int, hash string) bool {

	processedMessageMap := d.getProcessedMessageMap(round)
	chunkHashString := string(hash)
	_, ok := processedMessageMap[chunkHashString]
	return ok
}

func (d *Demux) markAsProcessed(round int, hash string) {

	processedMessageMap := d.getProcessedMessageMap(round)
	processedMessageMap[hash] = struct{}{}
}

func (d *Demux) getCorrespondingBlockChunkChan(round int) chan Chunk {

	if val, ok := d.blockChunkChanMap[round]; ok {
		return val
	}

	chunkChan := make(chan Chunk, channelCapacity)
	d.blockChunkChanMap[round] = chunkChan

	return chunkChan
}