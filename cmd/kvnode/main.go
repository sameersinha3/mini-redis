package main

import (
    "distributed-kv-store-go/api"
    "distributed-kv-store-go/internal/cluster"
    "distributed-kv-store-go/internal/kv"
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
    self := fmt.Sprintf("http://localhost:%s", port)

    store := kv.NewStore()
    cluster := cluster.NewPeerManager(self, peers)
    cluster.ElectLeader()
    handler := api.NewHandler(store, cluster)

    log.Printf("Starting server at :%s...", port)
    log.Fatal(http.ListenAndServe(":"+port, handler.Router()))
}