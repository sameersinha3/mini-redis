package raft

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (rn *RaftNode) HeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	termStr := r.Header.Get("X-Term")
	term, err := strconv.Atoi(termStr)
	if err != nil {
		http.Error(w, "Invalid term", http.StatusBadRequest)
		return
	}
	rn.HandleHeartbeat(term)
	w.WriteHeader(http.StatusOK)
}

func (rn *RaftNode) VoteRequestHandler(w http.ResponseWriter, r *http.Request) {
	type VoteRequest struct {
		Term        int    `json:"term"`
		CandidateID string `json:"candidateID"`
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid vote request", http.StatusBadRequest)
		return
	}

	granted := rn.handleVoteRequest(req.Term, req.CandidateID)
	resp := map[string]bool{"voteGranted": granted}
	json.NewEncoder(w).Encode(resp)
}

func (rn *RaftNode) AppendEntriesHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (rn *RaftNode) GetLeaderHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"leader": rn.GetLeader()}
	json.NewEncoder(w).Encode(resp)
}
