package internal

import (
	"fmt"
	"log"
)

type EventType int

const (
	FirstChunkReceived EventType = iota
	MessageReceived
	NetworkUsage
	DisseminationFailure
)

func (e EventType) String() string {
	switch e {
	case FirstChunkReceived:
		return "FIRST CHUNK RECEIVED"
	case MessageReceived:
		return "MESSAGE RECEIVED"
	case NetworkUsage:
		return "NETWORK USAGE"
	case DisseminationFailure:
		return "DISSEMINATION FAILURE"
	default:
		panic(fmt.Errorf("undefined enum value %d", e))
	}
}

type Event struct {
	Round        int
	Type         EventType
	ElapsedTime  int
	NetworkUsage int64
}

type StateLogger struct {
	nodeID int
	events []Event
}

func NewStateLogger(nodeID *int) *StateLogger {
	return &StateLogger{
		nodeID: *nodeID,
	}
}

func (s *StateLogger) FirstChunkReceived(round int, elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, round, "FIRST_CHUNK_RECEIVED", elapsedTime)
	s.events = append(s.events, Event{Round: round, Type: FirstChunkReceived, ElapsedTime: int(elapsedTime)})
}

func (s *StateLogger) MessageReceived(round int, elapsedTime int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, round, "MESSAGE_RECEIVED", elapsedTime)
	s.events = append(s.events, Event{Round: round, Type: MessageReceived, ElapsedTime: int(elapsedTime)})
}

func (s *StateLogger) NetworkUsage(round int, usage int64) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, round, "NETWORK_USAGE", usage)
	s.events = append(s.events, Event{Round: round, Type: NetworkUsage, NetworkUsage: usage})
}

func (s *StateLogger) DisseminationFailure(round int, leader int) {
	log.Printf("stats\t%d\t%d\t%s\t%d\t", s.nodeID, round, "DISSEMINATION_FAILURE", leader)
	s.events = append(s.events, Event{Round: round, Type: DisseminationFailure, ElapsedTime: leader})
}

func (s *StateLogger) GetEvents() []Event {
	return s.events
}
