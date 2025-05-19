package main

import (
    "distributed-kv-store-go/api"
    "distributed-kv-store-go/internal/cluster"
    "distributed-kv-store-go/internal/kv"
    "distributed-kv-store-go/internal/raft"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    peers := strings.Split(os.Getenv("PEERS"), ",")
    self := fmt.Sprintf("localhost:%s", port)

    store := kv.NewStore()
    cluster := cluster.NewPeerManager(self, peers)

    rn := raft.NewRaftNode(self, peers)

    handler := api.NewHandler(store, cluster)

    mux := http.NewServeMux()
    raft.RegisterRaftHandlers(mux, rn)

    mux.Handle("/api/", handler.Router())

    log.Printf("Starting server at :%s...", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
