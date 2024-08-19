package registry

import (
	"crypto/sha256"
	"fmt"
)

type NodeConfig struct {
	NodeCount                int
	EpochSeed                []byte
	EndRound                 int
	RoundSleepTime           int
	GossipFanout             int
	ConnectionCount          int
	SourceCount              int
	MessageSize              int
	MessageChunkCount        int
	DataChunkCount           int
	EndOfExperimentSleepTime int
	FaultyNodePercent        int
	DisseminationTimeout     int
}

func (n NodeConfig) Hash() []byte {
	str := fmt.Sprintf("%d,%x,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d", n.NodeCount, n.EpochSeed, n.EndRound, n.RoundSleepTime, n.GossipFanout, n.ConnectionCount, n.SourceCount, n.MessageSize, n.MessageChunkCount, n.DataChunkCount, n.EndOfExperimentSleepTime, n.FaultyNodePercent, n.DisseminationTimeout)

	hash := sha256.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	return hash.Sum(nil)
}

func (n *NodeConfig) CopyFields(cp NodeConfig){
	n.NodeCount = cp.NodeCount
	n.EpochSeed = n.EpochSeed[:0]
	n.EpochSeed = append(n.EpochSeed, cp.EpochSeed...)
	n.EndRound = cp.EndRound
	n.RoundSleepTime = cp.RoundSleepTime
	n.GossipFanout = cp.GossipFanout
	n.ConnectionCount = cp.ConnectionCount
	n.SourceCount = cp.SourceCount
	n.MessageSize = cp.MessageSize
	n.MessageChunkCount = cp.MessageChunkCount
	n.DataChunkCount = cp.DataChunkCount
	n.EndOfExperimentSleepTime = cp.EndOfExperimentSleepTime
	n.FaultyNodePercent = cp.FaultyNodePercent
	n.DisseminationTimeout = cp.DisseminationTimeout
}

