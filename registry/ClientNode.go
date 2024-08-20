package registry

import (
	"EncrypteDL/Gossip-Based-Systems/internal"
	"net/rpc"
)

type RegistryClient struct {
	rpcClient rpc.Client
	nodeInfo NodeInfo
}

func NewRegistryClient(registryAdress string, currentNodeInfo NodeInfo) RegistryClient{
	rpcClient, err := rpc.Dial("TCP", registryAdress)
	if err != nil{
		panic(err)
	}

	registeryClient := RegistryClient{
		rpcClient: *rpcClient,
		nodeInfo: currentNodeInfo,
	}

	return registeryClient
}

/*
RegisterNode()
GetConfig()
GetNodeList()
UploadStats()
NodeStarted()
NodeFailed()
NodeFinished()
*/

//RegisterNode register a node and returns assigned node ID 
func(r RegistryClient) RegisterNode() int{
	err := r.rpcClient.Call("Noderegistry.Rigister", r.nodeInfo, &r.nodeInfo)
	if err != nil{
		panic(err)
	}

	return r.nodeInfo.ID
}

func(r RegistryClient) GetConfig() NodeConfig{
	config := NodeConfig{}

	err := r.rpcClient.Call("Noderegistry.GetConfig", r.nodeInfo, &config)
	if err != nil{
		panic(err)
	}

	return config
}

func (r RegistryClient) GetNodeList() []NodeInfo {

	nodeList := NodeList{}
	err := r.rpcClient.Call("NodeRegistry.GetNodeList", r.nodeInfo, &nodeList)
	if err != nil {
		panic(err)
	}

	return nodeList.Nodes
}

func (r RegistryClient) UploadStats(statList internal.StatList) {

	err := r.rpcClient.Call("NodeRegistry.UploadStats", statList, nil)
	if err != nil {
		panic(err)
	}
}

func (r RegistryClient) NodeStarted() {

	err := r.rpcClient.Call("NodeRegistry.NodeStarted", r.nodeInfo, nil)
	if err != nil {
		panic(err)
	}
}

func (r RegistryClient) NodeFailed() {

	err := r.rpcClient.Call("NodeRegistry.NodeFailed", r.nodeInfo, nil)
	if err != nil {
		panic(err)
	}
}

func (r RegistryClient) NodeFinished() {

	err := r.rpcClient.Call("NodeRegistry.NodeFinished", r.nodeInfo, nil)
	if err != nil {
		panic(err)
	}
}