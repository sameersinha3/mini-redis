package raft

import (
	"net/http"
)

func RegisterRaftHandlers(mux *http.ServeMux, rn *RaftNode) {
	mux.HandleFunc("/raft/heartbeat", rn.HeartbeatHandler)
	mux.HandleFunc("/raft/vote", rn.VoteRequestHandler)
	mux.HandleFunc("/raft/appendentries", rn.AppendEntriesHandler)
	mux.HandleFunc("/raft/leader", rn.GetLeaderHandler)
}
