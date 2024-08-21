package dissemination

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"EncrypteDL/Gossip-Based-Systems/server"
	"fmt"
	"log"
	"time"
)

func receiveMultipleBlocks(round int, demux *internal.Demux, chunkCount int, dataChunkCount int, peerSet *server.PeerSet, leaderCount int, statLogger *internal.StateLogger, electedLeaders []int, timeout int) ([]internal.Message, []int) {

	chunkChan, err := demux.GetMessageChunkChan(round)
	if err != nil {
		panic(err)
	}

	receiver := newBlockReceiver(leaderCount, chunkCount, dataChunkCount)
	firstChunkReceived := false

	var timeOutChan <-chan time.Time

	// if timeout value is equal or smaller to 0, the chanel will be nil, and
	// all operations will block forever except close operation that panics
	if timeout > 0 {
		timeOutChan = time.After(time.Duration(timeout) * time.Second)
	}

	chunkReceivedMap := make(map[int]bool)
	for _, leader := range electedLeaders {
		chunkReceivedMap[leader] = false
	}

	for !receiver.ReceivedAll() {

		select {
		case c := <-chunkChan:
			{

				if c.Round != round {
					panic(fmt.Errorf("expected round is %d and chunk from round %d", round, c.Round))
				}

				//TODO: considers only one leader
				if c.Issuer != electedLeaders[0] {
					log.Printf("A chunk is received from previous leader %d, discarding the chunk", c.Issuer)
					continue
				}

				// a chunk is received from the leader
				chunkReceivedMap[c.Issuer] = true

				if !firstChunkReceived {
					statLogger.FirstChunkReceived(round, (time.Now().UnixMilli() - c.Time))
					firstChunkReceived = true
				}

				receiver.AddChunk(c)
				peerSet.ForwardChunk(c)
			}
		case <-timeOutChan:
			// // checks for unresponsive leaders
			// var leadersToRemove []int
			// for leader, value := range chunkReceivedMap {
			// 	if value == false {
			// 		leadersToRemove = append(leadersToRemove, leader)
			// 	}
			// }

			// // this means that at least one leader is unresponsive
			// if len(leadersToRemove) > 0 {
			// 	return nil, leadersToRemove
			// }

			// WARNING: all leaders are evicted...
			log.Printf("Dissemination Timeout expired: All leaders considered FAULTY!!!")
			return nil, electedLeaders
		}

	}

	return receiver.GetBlocks(), nil
}
