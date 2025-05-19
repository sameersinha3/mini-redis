package api

import (
    "distributed-kv-store-go/internal/cluster"
    "distributed-kv-store-go/internal/kv"
    "distributed-kv-store-go/internal/raft"
    "encoding/json"
    "net/http"
    

    "github.com/gorilla/mux"
)

type Handler struct {
    store  *kv.Store
    peers  *cluster.PeerManager
    router *mux.Router
    raftNode *raft.RaftNode
}

func NewHandler(store *kv.Store, peers *cluster.PeerManager, rn *raft.RaftNode) *Handler {
    h := &Handler{store: store, peers: peers, raftNode: rn}
    r := mux.NewRouter()
    r.HandleFunc("/set", h.SetHandler).Methods("POST")
    r.HandleFunc("/get/{key}", h.GetHandler).Methods("GET")
    r.HandleFunc("/delete/{key}", h.DeleteHandler).Methods("DELETE")
    r.HandleFunc("/replicate", h.ReplicationHandler).Methods("POST")
    r.HandleFunc("/replicate/delete", h.ReplicateDeleteHandler).Methods("POST")
    r.HandleFunc("/leader", h.LeaderHandler).Methods("GET")
    h.router = r
    return h
}

func (h *Handler) Router() http.Handler {
    return h.router
}

func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
    if h.raftNode.GetLeader() != h.peers.Self {
        http.Error(w, "Not leader", http.StatusForbidden)
        return
    }
    var data map[string]string
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    key := data["key"]
    value := data["value"]
    if key == "" {
        http.Error(w, "Key is required", http.StatusBadRequest)
        return
    }
    h.store.Set(key, value)
    h.peers.Replicate(key, value)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    key := vars["key"]
    val, ok := h.store.Get(key)
    if !ok {
        http.NotFound(w, r)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"value": val})
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    key := vars["key"]
    
    if deleted := h.store.Delete(key); !deleted {
        http.NotFound(w, r)
        return
    }
    
    // Notify peers about the deletion
    h.peers.ReplicateDelete(key)
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *Handler) ReplicationHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    key := data["key"]
    value := data["value"]
    h.store.Set(key, value)
    w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReplicateDeleteHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    key := data["key"]
    h.store.Delete(key)
    w.WriteHeader(http.StatusOK)
}

func (h *Handler) LeaderHandler(w http.ResponseWriter, r *http.Request) {
    leader := h.raftNode.GetLeader()
    json.NewEncoder(w).Encode(map[string]string{"leader": leader})
}