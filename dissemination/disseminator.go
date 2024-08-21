package dissemination

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"EncrypteDL/Gossip-Based-Systems/registry"
	"EncrypteDL/Gossip-Based-Systems/server"
	"errors"
	"time"
)

// BlockNotValid is returned if the block can not pass vaslidity test
var ErrBlockNotValid = errors.New("received block is not valid")

type Disseminator struct {
	demultiplexer *internal.Demux
	nodeConfig    registry.NodeConfig
	peerSet       server.PeerSet
	statLogger    *internal.StateLogger
}

func NewDisseminator(demux *internal.Demux, config registry.NodeConfig, peerSet server.PeerSet, statLogger *internal.StateLogger) *Disseminator {

	d := &Disseminator{
		demultiplexer: demux,
		nodeConfig:    config,
		peerSet:       peerSet,
		statLogger:    statLogger,
	}

	return d
}

func (d *Disseminator) SubmitMessage(round int, message internal.Message) {

	// sets the round for demultiplexer
	d.demultiplexer.UpdateRound(round)

	// chunks the block
	chunks := internal.ChunkMessage(message, d.nodeConfig.MessageChunkCount, d.nodeConfig.DataChunkCount)
	//log.Printf("proposing block %x\n", encodeBase64(merkleRoot[:15]))

	// disseminate chunks over different nodes
	d.peerSet.DissaminateChunks(chunks)

	//return d.WaitForMessage(round)
}

func (d *Disseminator) WaitForMessage(round int, electedLeaders []int, timeout int) ([]internal.Message, []int) {

	// sets the round for demultiplexer
	d.demultiplexer.UpdateRound(round)

	messages, leadersToRemove := receiveMultipleBlocks(round, d.demultiplexer, d.nodeConfig.MessageChunkCount, d.nodeConfig.DataChunkCount, &d.peerSet, d.nodeConfig.SourceCount, d.statLogger, electedLeaders, timeout)

	if leadersToRemove != nil {
		//TODO: considers single leader
		d.statLogger.DisseminationFailure(round, leadersToRemove[0])
		return nil, leadersToRemove
	}

	var maxElapsedTime int64
	for i := range messages {
		elapsedTime := time.Now().UnixMilli() - messages[i].Time
		if elapsedTime > maxElapsedTime {
			maxElapsedTime = elapsedTime
		}
	}
	d.statLogger.MessageReceived(round, (maxElapsedTime))

	return messages, nil
}
