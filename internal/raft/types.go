package raft

import (
	"sync"
)

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

type RaftNode struct {
	mu                 sync.Mutex
	selfID             string
	peers              []string
	state              State
	currentTerm        int
	votedFor           string
	leaderID           string
	electionResetEvent chan struct{}
	lastHeartbeat      int64
}
