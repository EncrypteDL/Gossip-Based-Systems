package main

import (
	"EncrypteDL/Gossip-Based-Systems/dissemination"
	"EncrypteDL/Gossip-Based-Systems/internal"
	"EncrypteDL/Gossip-Based-Systems/registry"
	"EncrypteDL/Gossip-Based-Systems/server"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

func main() {

	hostName := getEnvWithDefault("NODE_HOSTENAME", "127.0.0.1")
	registryAddress := getEnvWithDefault("REGISTRY_ADDRESS", "localhost:1234")
	processIndex := getEnvWithDefault("PROCESS_INDEX", "-1")

	log.Printf("Process Index: %s\n", &processIndex)
	log.Printf("Host Name: %s\n", &hostName)
	log.Printf("registry address: %s\n", &registryAddress)

	demux := internal.NewDemultiplexer(0)
	server := server.NewServer(demux)

	err := rpc.Register(server)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprint("%s:", hostName))
	if err != nil {
		log.Fatal("listen error:", e)
	}

	//Start the server
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

	//Node failed
	defer func() {
		if r := recover(); r != nil {
			log.Println("Failed")
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

	peerSet := createdPeerSet(nodeList, nodeConfig.GossipFanout, nodeInfo.ID, nodeInfo.IPAddress, nodeConfig.ConnectionCount)
	statLogger := internal.NewStateLogger(nodeInfo.ID)
	rapidChain := dissemination.NewDisseminator(demux, nodeConfig, peerSet, statLogger)

	//Node started
	registry.NodeStarted()

	runConsensus(rapidChain, nodeConfig.EndRound, nodeConfig.RoundSleepTime, nodeInfo.ID, nodeConfig.SourceCount, nodeConfig.MessageSize, nodeList, nodeConfig.DisseminationTimeout)

	sleepTime := time.Duration(nodeConfig.EndOfExperimentSleepTime) * time.Second
	log.Printf("Reached target round count. Shutting down %s/n", sleepTime)
	time.Sleep(sleepTime)

	log.Printf("Getting network usage...\n")
	bandWithUsage := getBandwitchUsage(processIndex)
	statLogger.NetworkUsage(-1, bandWithUsage)

	//Collect stats abd uploads to registry 
	log.Printf("Uploading stats to the registry\n")
	events :=statLogger.GetEvents()
	statList :=  internal.StatList{
		IPAddress: nodeInfo.IPAddress,
		PortNumber: nodeInfo.PortNumber,
		NodeID: nodeInfo.ID,
		Events: events,
	}
	registry.UploadStats(statList)

	//Node finished 
	log.Printf("existing as expected...\n")
}

