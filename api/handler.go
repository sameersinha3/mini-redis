package api

import (
    "distributed-kv-store-go/internal/cluster"
    "distributed-kv-store-go/internal/kv"
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
)

type Handler struct {
    store  *kv.Store
    peers  *cluster.PeerManager
    router *mux.Router
}

func NewHandler(store *kv.Store, peers *cluster.PeerManager) *Handler {
    h := &Handler{store: store, peers: peers}
    r := mux.NewRouter()
    r.HandleFunc("/set", h.SetHandler).Methods("POST")
    r.HandleFunc("/get/{key}", h.GetHandler).Methods("GET")
    r.HandleFunc("/replicate", h.ReplicationHandler).Methods("POST")
    h.router = r
    return h
}

func (h *Handler) Router() http.Handler {
    return h.router
}

func (h *Handler) SetHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    json.NewDecoder(r.Body).Decode(&data)
    key := data["key"]
    value := data["value"]
    h.store.Set(key, value)
    h.peers.Replicate(key, value)
    w.WriteHeader(http.StatusOK)
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

func (h *Handler) ReplicationHandler(w http.ResponseWriter, r *http.Request) {
    var data map[string]string
    json.NewDecoder(r.Body).Decode(&data)
    key := data["key"]
    value := data["value"]
    h.store.Set(key, value)
    w.WriteHeader(http.StatusOK)
}