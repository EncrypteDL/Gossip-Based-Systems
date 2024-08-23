package main

import (
	"EncrypteDL/Gossip-Based-Systems/dissemination"
	"EncrypteDL/Gossip-Based-Systems/internal"
	"EncrypteDL/Gossip-Based-Systems/registry"
	"EncrypteDL/Gossip-Based-Systems/server"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {

	hostname := getEnvWithDefault("NODE_HOSTNAME", "127.0.0.1")
	registryAddress := getEnvWithDefault("REGISTRY_ADDRESS", "localhost:1234")
	processIndex := getEnvWithDefault("PROCESS_INDEX", "-1")

	log.Printf("Process Index: %s\n", processIndex)

	demux := internal.NewDemultiplexer(0)
	server := server.NewServer(demux)

	err := rpc.Register(server)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf("%s:", hostname))
	if e != nil {
		log.Fatal("listen error:", e)
	}

	// start serving
	go func() {
		for {
			conn, _ := l.Accept()
			go func() {
				rpc.ServeConn(conn)
			}()
		}
	}()

	log.Printf("p2p server started on %s\n", l.Addr().String())
	nodeInfo := getNodeInfo(l.Addr().String())

	registry := registry.NewRegistryClient(registryAddress, nodeInfo)

	nodeInfo.ID = registry.RegisterNode()
	log.Printf("node registeration successful, assigned ID is %d\n", nodeInfo.ID)

	nodeConfig := registry.GetConfig()

	var nodeList []registry.NodeInfo

	///// Node Failed /////
	defer func() {
		if r := recover(); r != nil {
			log.Println("########### FAILED ############")
			registry.NodeFailed()
		}
	}()

	for {
		nodeList = registry.GetNodeList()
		nodeCount := len(nodeList)
		if nodeCount == nodeConfig.NodeCount {
			break
		}
		time.Sleep(2 * time.Second)
		log.Printf("received node list %d/%d\n", nodeCount, nodeConfig.NodeCount)
	}

	peerSet := createdPeerSet(nodeList, nodeConfig.GossipFanout, nodeInfo.ID, nodeInfo.IPAdrees, nodeConfig.ConnectionCount)
	statLogger := internal.NewStateLogger(&nodeInfo.ID)
	rapidchain := dissemination.NewDisseminator(demux, nodeConfig, peerSet, statLogger)

	///// Node Started /////
	registry.NodeStarted()

	runConsensus(rapidchain, nodeConfig.EndRound, nodeConfig.RoundSleepTime, nodeInfo.ID, nodeConfig.SourceCount, nodeConfig.MessageSize, nodeList, nodeConfig.DisseminationTimeout)

	sleepTime := time.Duration(nodeConfig.EndOfExperimentSleepTime) * time.Second
	log.Printf("Reached target round count. Shutting down in %s\n", sleepTime)
	time.Sleep(sleepTime)

	log.Printf("getting network usage...\n")
	bandwidthUsage := getBandwitchUsage(processIndex)
	statLogger.NetworkUsage(-1, bandwidthUsage)

	// collects stats abd uploads to registry
	log.Printf("uploading stats to the registry\n")
	events := statLogger.GetEvents()
	statList := internal.StatList{IPAddress: nodeInfo.IPAdrees, PortNumber: nodeInfo.PortNumber, NodeID: nodeInfo.ID, Events: events}
	registry.UploadStats(statList)

	///// Node Finished /////
	registry.NodeFinished()

	log.Printf("exiting as expected...\n")
}

func getBandwitchUsage(processIndex string) int64 {
	cmd := exec.Command("nin/bash", "./get-network-usage.sh", processIndex)
	output, err := cmd.Output()

	if err != nil {
		log.Printf("error accured while executing get-newtrok-usage.sh %s\n", err)
		return 0
	}

	outputString := strings.TrimSpace(string(output))
	bandwitchUsage, err := strconv.ParseInt(outputString, 10, 64)
	if err != nil {
		log.Printf("error occured while converting %s to int64 %s\n", outputString, err)
		return 0
	}
	return bandwitchUsage
}

func createdPeerSet(nodeList []registry.NodeInfo, fanOut int, nodeID int, localIpAddress string, connectionCount int) server.PeerSet {
	var copyNodeList []registry.NodeInfo
	copyNodeList = append(copyNodeList, nodeList...)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copyNodeList), func(i, j int) { copyNodeList[i], copyNodeList[j] = copyNodeList[j], copyNodeList[i] })

	peerSet := server.PeerSet{}

	peerCount := 0
	for i := 0; i < len(copyNodeList); i++ {
		if peerCount == fanOut {
			break
		}

		peer := copyNodeList[i]
		//TODO: do not connect nodes from local machine
		if peer.ID == nodeID || peer.IPAdrees == localIpAddress {
			continue
		}

		err := peerSet.AddPeer(peer.IPAdrees, peer.PortNumber, connectionCount)
		if err != nil {
			panic(err)
		}
		log.Printf("new peer added: %s:%d ID %d\n", peer.IPAdrees, peer.PortNumber, peer.ID)
		peerCount++
	}

	return peerSet

}

func getNodeInfo(netAddress string) registry.NodeInfo {
	tokens := strings.Split(netAddress, ":")

	ipAddress := tokens[0]
	portNumber, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}

	return registry.NodeInfo{
		IPAdrees:   ipAddress,
		PortNumber: portNumber,
	}
}

func runConsensus(rc *dissemination.Disseminator, numberOfRounds int, roundSleepTime int, nodeID int, leaderCount int, blockSize int, nodeList []registry.NodeInfo, timeout int) {

	currentRound := 1
	for currentRound <= numberOfRounds {

		log.Printf("+++++++++ Round %d +++++++++++++++\n", currentRound)

		var messages []internal.Message

		// if elected as a leader submits a block
		isElected, electedLeaders := isElectedAsLeader(nodeList, currentRound, nodeID, leaderCount)

		if isElected {
			log.Println("elected as leader")
			b := createBlock(currentRound, nodeID, blockSize, leaderCount)
			rc.SubmitMessage(currentRound, b)
		}

		// TODO: is it better to log individual messages?
		// waits to deliver the block
		log.Printf("waiting to deliver messages...\n")
		messages, unresponsiveLeaders := rc.WaitForMessage(currentRound, electedLeaders, timeout)

		if unresponsiveLeaders != nil {
			log.Printf("Unresponsive leader detected: %v\n", unresponsiveLeaders)
			nodeList = removeUnresponsiveLeaders(unresponsiveLeaders, nodeList)
			log.Printf("Unresponsive leaders are removed.")
			continue
		}

		log.Printf("all messages delivered.\n")
		payloadSize := 0
		for i := range messages {
			log.Printf("Round: %d Message[%d] %x\n", currentRound, i, internal.EncodeBase64(messages[i].Hash())[:15])
			payloadSize += len(messages[i].Payload)
		}

		log.Printf("round finished, payload size payload size: %d bytes\n", payloadSize)

		currentRound++

		sleepTime := int64(roundSleepTime*1000) - (time.Now().UnixMilli() - messages[0].Time)
		log.Printf("sleeping for %d ms\n", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)

		//sleepTime := time.Duration(roundSleepTime) * time.Second
		//log.Printf("sleeping for %s\n", sleepTime)
		//time.Sleep(sleepTime)

	}

}
