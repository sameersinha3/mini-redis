package raft

import (
	"encoding/gob"
	"os"
)

type Snapshot struct {
	State             map[string]string `json:"state"`
	LastIncludedIndex int               `json:"lastIncludedIndex"`
	LastIncludedTerm  int               `json:"lastIncludedTerm"`
}

func (r *RaftNode) SaveSnapshot(state map[string]string, index int, term int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snapshot := Snapshot{
		State:             state,
		LastIncludedIndex: index,
		LastIncludedTerm:  term,
	}

	file, err := os.Create("snapshot.gob")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(snapshot)
}

func (r *RaftNode) LoadSnapshot() (*Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Open("snapshot.gob")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var snapshot Snapshot
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, err
	}

	r.snapshotIndex = snapshot.LastIncludedIndex
	r.snapshotTerm = snapshot.LastIncludedTerm
	return &snapshot, nil
}
