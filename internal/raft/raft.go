package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"bytes"
	"encoding/json"
	"distributed-kv-store-go/internal/kv"
)



func NewRaftNode(selfID string, peers []string, store *kv.Store) *RaftNode {
	rn := &RaftNode{
		selfID:             selfID,
		peers:              peers,
		store:				store,
		state:              Follower,
		currentTerm:        0,
		votedFor:           "",
		electionResetEvent: make(chan struct{}),
	}
	rn.LoadSnapshot()
	if rn.snapshot != nil {
		rn.store.LoadFromSnapshot(rn.snapshot.State)
	}
	go rn.electionLoop()
	return rn
}

func (r *RaftNode) electionLoop() {
	for {
		timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
		timer := time.NewTimer(timeout)
		select {
		case <-r.electionResetEvent:
			timer.Stop()
		case <-timer.C:
			r.startElection()
		}
	}
}

func (r *RaftNode) startElection() {
	r.mu.Lock()
	r.state = Candidate
	r.currentTerm++
	r.votedFor = r.selfID
	term := r.currentTerm
	r.mu.Unlock()

	votes := 1
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, peer := range r.peers {
		if peer == r.selfID {
			continue
		}
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()
			ok := sendVoteRequest(peer, r.selfID, term)
			if ok {
				mu.Lock()
				votes++
				mu.Unlock()
			}
		}(peer)
	}
	wg.Wait()

	if votes > len(r.peers)/2 {
		r.mu.Lock()
		r.state = Leader
		r.leaderID = r.selfID
		r.mu.Unlock()
		log.Printf("%s became leader for term %d", r.selfID, term)
		r.startHeartbeatLoop()
	}
}

func (r *RaftNode) startHeartbeatLoop() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			r.mu.Lock()
			if r.state != Leader {
				r.mu.Unlock()
				return
			}
			term := r.currentTerm
			r.mu.Unlock()

			for _, peer := range r.peers {
				if peer == r.selfID {
					continue
				}
				go func(peer string) {
					url := "http://" + peer + "/raft/heartbeat"
					req, _ := http.NewRequest("POST", url, nil)
					req.Header.Set("X-Term", fmt.Sprintf("%d", term))
					_, err := http.DefaultClient.Do(req)
					if err != nil {
						log.Printf("Failed to send heartbeat to %s: %v", peer, err)
					}
				}(peer)
			}
			<-ticker.C
		}
	}()
}

func (r *RaftNode) HandleHeartbeat(term int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if term >= r.currentTerm {
		r.state = Follower
		r.currentTerm = term
		select {
		case r.electionResetEvent <- struct{}{}:
		default:
		}
	}
}

func (rn *RaftNode) handleVoteRequest(term int, candidateID string) bool {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if term < rn.currentTerm {
		return false
	}

	if (rn.votedFor == "" || rn.votedFor == candidateID) && term >= rn.currentTerm {
		rn.votedFor = candidateID
		rn.currentTerm = term
		select {
		case rn.electionResetEvent <- struct{}{}:
		default:
		}
		return true
	}

	return false
}

func (rn *RaftNode) GetLeader() string {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.leaderID
}

func sendVoteRequest(peer string, candidateID string, term int) bool {
	url := fmt.Sprintf("http://%s/raft/vote", peer)

	reqBody := map[string]interface{}{
		"term":        term,
		"candidateID": candidateID,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var respBody struct {
		VoteGranted bool `json:"voteGranted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return false
	}

	return respBody.VoteGranted
}

func (r *RaftNode) IsLeader() bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.state == Leader
}

func (r *RaftNode) LastApplied() int {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.lastApplied
}

func (r *RaftNode) CurrentTerm() int {
    r.mu.Lock()
    defer r.mu.Unlock()
    return r.currentTerm
}