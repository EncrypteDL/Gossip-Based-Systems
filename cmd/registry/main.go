package main

import (
	"EncrypteDL/Gossip-Based-Systems/registry"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
)

// func readConfigFromFile() registry.NodeConfig {
// 	data, err := ioutil.ReadFile(configFile)
// 	if err != nil {
// 		panic(err)
// 	}

// 	config := registry.NodeConfig{}
// 	json.Marshal(data, &config)

// 	return config
// }

func main() {
	statusLogger := registry.NewStatutsLogger()
	defer func() {
		if r := recover(); r != nil {
			statusLogger.LogFailed()
		}
	}()

	nodeConfig := readConfigFromFile()
	nodeRegistry := registry.NewNodeRegistry(nodeConfig, statusLogger)

	err := rpc.Register(nodeRegistry)
	if err != nil {
		panic(err)
	}

	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error:", e)
	}

	log.Printf("registery service started and listening on :1234\n")

	for {
		conn, _ := l.Accept()
		go func() {
			rpc.ServeConn(conn)
		}()
	}
}

func readConfigFromFile() registry.NodeConfig {

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	config := registery.NodeConfig{}
	json.Unmarshal(data, &config)

	return config
}
