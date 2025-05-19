package main

import (
    "distributed-kv-store-go/api"
    "distributed-kv-store-go/internal/cluster"
    "distributed-kv-store-go/internal/kv"
    "distributed-kv-store-go/internal/raft"
    "time"
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

    rn := raft.NewRaftNode(self, peers, store)
    go func() {
        for {
            time.Sleep(10 * time.Second)
            if rn.IsLeader() {
                snapshotState := store.GetAll()
                if err := rn.SaveSnapshot(snapshotState, rn.LastApplied(), rn.CurrentTerm()); err != nil {
                    log.Printf("Error saving snapshot: %v", err)
                } else {
                    log.Println("Snapshot saved successfully")
                }
            }
        }
    }()
    

    handler := api.NewHandler(store, cluster, rn)

    mux := http.NewServeMux()
    raft.RegisterRaftHandlers(mux, rn)

    mux.Handle("/api/", http.StripPrefix("/api", handler.Router()))


    log.Printf("Starting server at :%s...", port)
    log.Fatal(http.ListenAndServe(":"+port, mux))
}
