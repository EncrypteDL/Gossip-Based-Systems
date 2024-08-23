package main

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"EncrypteDL/Gossip-Based-Systems/registry"
	"log"
	"math/rand"
	"os"
	"time"
)

func removeUnresponsiveLeaders(unresponsiveLeaders []int, nodeList []registry.NodeInfo) []registry.NodeInfo {

	var newNodeList []registry.NodeInfo
	for _, nodeInfo := range nodeList {

		isInList := false
		for _, l := range unresponsiveLeaders {
			if l == nodeInfo.ID {
				isInList = true
				break
			}
		}

		if isInList == false {
			newNodeList = append(newNodeList, nodeInfo)
		}

	}

	return newNodeList
}

func createBlock(round int, nodeID int, blockSize int, leaderCount int) internal.Message {

	payloadSize := blockSize

	block := internal.Message{
		Round:   round,
		Issuer:  nodeID,
		Time:    time.Now().UnixMilli(),
		Payload: internal.GetRandomBySlices(payloadSize),
	}

	return block
}

func getEnvWithDefault(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		val = defaultValue
	}

	log.Printf("%s=%s\n", key, val)
	return val
}

func isElectedAsLeader(nodeList []internal.NodeInfo, round int, nodeID int, leaderCount int) (bool, []int) {

	// assumes that node list is same for all nodes
	// shuffle the node list using round number as the source of randomness
	rand.Seed(int64(round))
	rand.Shuffle(len(nodeList), func(i, j int) { nodeList[i], nodeList[j] = nodeList[j], nodeList[i] })

	var electedLeaders []int
	isElected := false
	for i := 0; i < leaderCount; i++ {
		electedLeaders = append(electedLeaders, nodeList[i].ID)
		if nodeList[i].ID == nodeID {
			log.Println("=== elects as a leader ===")
			isElected = true
		}
	}

	log.Printf("Elected leaders: %v\n", electedLeaders)

	return isElected, electedLeaders
}
