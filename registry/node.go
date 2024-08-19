package registry

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

// NodeInfo keeps node info
type NodeInfo struct {
	ID         int
	IPAdrees   string
	PortNumber int
}

type NodeList struct {
	Nodes []NodeInfo
}

type NodeRegistry struct {
	mutex         sync.Mutex
	registerNodes []NodeInfo

	failedNodes   []NodeInfo
	finishedNodes []NodeInfo
	startedNodes  []NodeInfo

	config       NodeConfig
	statKeeper   *StatKeeper
	statusLogger *StatusLogger
}

func NewNodeRegistry(config NodeConfig, statusLogger *StatusLogger) *NodeRegistry {
	return &NodeRegistry{
		config:       config,
		statusLogger: statusLogger,
	}
}

// Register registers a node with specific node info
func (n *NodeRegistry) Register(nodeInfo *NodeInfo, reply *NodeInfo) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	//assigns a node ID. smallest node ID
	nodeID := len(n.registerNodes) + 1
	nodeInfo.ID = nodeID

	n.registerNodes = append(n.registerNodes, *nodeInfo)
	log.Println("new node registerd; ip address %s port numnber %d, registered node count: %d\n", nodeInfo.IPAdrees, nodeInfo.PortNumber, len(n.registerNodes))

	reply.IPAdrees = nodeInfo.IPAdrees
	reply.PortNumber = nodeInfo.PortNumber
	reply.ID = nodeInfo.ID

	return nil
}

/* //TODO:
- Register
- NodeStarted
- NodeFailed
- NodeFinished
- Unregister
- GetConfig
- GetNodelist
- UploadStats
*/

func (n *NodeRegistry) NodeStarted(nodeInfor *NodeInfo, reply *int) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	log.Printf("node stated %d\n", nodeInfor.ID)

	startedNodeCount := len(n.startedNodes)
	failedNodeConnect := len(n.failedNodes)

	if (startedNodeCount + failedNodeConnect) == n.config.NodeCount {
		n.statusLogger.LogStarted()
	} else {
		log.Printf("%d failed %d sterted\n", failedNodeConnect, startedNodeCount)
	}
	return nil
}

func (n *NodeRegistry) NodeFailed(nodeInfo *NodeInfo, reply *int) error {

	n.mutex.Lock()
	defer n.mutex.Unlock()

	log.Printf("node failed %d\n", nodeInfo.ID)

	n.failedNodes = append(n.failedNodes, *nodeInfo)

	failedNodeCount := len(n.failedNodes)

	if failedNodeCount >= 20 {
		n.statusLogger.LogFailed()
		panic(fmt.Errorf("more than 20 nodes failed there must be a problem"))
	}

	return nil
}

func (n *NodeRegistry) NodeFinished(nodeInfo *NodeInfo, reply *int) error {

	n.mutex.Lock()
	defer n.mutex.Unlock()

	log.Printf("node fiished %d\n", nodeInfo.ID)

	n.finishedNodes = append(n.finishedNodes, *nodeInfo)

	finishedNodeCount := len(n.finishedNodes)
	failedNodeCount := len(n.failedNodes)

	if (finishedNodeCount + failedNodeCount) == n.config.NodeCount {
		n.statusLogger.LogFinished()
	} else {
		log.Printf("%d failed %d finished\n", failedNodeCount, finishedNodeCount)
	}

	return nil
}

func (nr *NodeRegistry) Unregister(remoteAddress string) {
	addressParts := strings.Split(remoteAddress, ":")

	if len(addressParts) != 2 {
		log.Printf("unknown address format, node couldnot unregistered %s \n", remoteAddress)
		return
	}

	ipAddress := addressParts[0]
	portNumber, err := strconv.Atoi(addressParts[1])
	if err != nil {
		log.Printf("could not parse the port number, error: %s, portnumber: %s\n", err, addressParts[1])
		return
	}

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nodeIndex := -1
	for i := range nr.registerNodes {
		if nr.registerNodes[i].IPAdrees == ipAddress && nr.registerNodes[i].PortNumber == portNumber {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		log.Printf("could not find %s in the registered node list to unregister\n", remoteAddress)
		return
	}

	nr.registerNodes = append(nr.registerNodes[:nodeIndex], nr.registerNodes[nodeIndex+1:]...)
	log.Printf("node %s unregistered successfully\n", remoteAddress)

}

// GetConfig is used to get config
func (nr *NodeRegistry) GetConfig(nodeInfo *NodeInfo, config *NodeConfig) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	config.CopyFields(nr.config)

	return nil
}

// GetNodeList returns node list
func (nr *NodeRegistry) GetNodeList(nodeInfo *NodeInfo, nodeList *NodeList) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	nodeList.Nodes = append(nodeList.Nodes, nr.registerNodes...)

	return nil
}

func (nr *NodeRegistry) UploadStats(stats *internal.StatList, reply *int) error {

	nr.mutex.Lock()
	defer nr.mutex.Unlock()

	log.Printf("node %d (%s:%d) uploading stats; event count %d \n", stats.NodeID, stats.IPAddress, stats.PortNumber, len(stats.Events))

	if nr.statKeeper == nil {
		nr.statKeeper = NewStatKeeper(nr.config)
	}

	nr.statKeeper.SaveStats(*stats)

	return nil
}
