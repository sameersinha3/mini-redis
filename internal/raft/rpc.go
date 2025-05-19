package raft

import (
	"net/http"
	"encoding/json"
)

func (rn *RaftNode) installSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	var snapshot Snapshot
	if err := json.NewDecoder(r.Body).Decode(&snapshot); err != nil {
		http.Error(w, "Invalid snapshot", http.StatusBadRequest)
		return
	}

	rn.mu.Lock()
	rn.snapshotIndex = snapshot.LastIncludedIndex
	rn.snapshotTerm = snapshot.LastIncludedTerm
	rn.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}


func RegisterRaftHandlers(mux *http.ServeMux, rn *RaftNode) {
    mux.HandleFunc("/raft/heartbeat", rn.HeartbeatHandler)
    mux.HandleFunc("/raft/vote", rn.VoteRequestHandler)
    mux.HandleFunc("/raft/appendentries", rn.AppendEntriesHandler)
    mux.HandleFunc("/raft/leader", rn.GetLeaderHandler)
    mux.HandleFunc("/raft/installSnapshot", rn.installSnapshotHandler)
}
