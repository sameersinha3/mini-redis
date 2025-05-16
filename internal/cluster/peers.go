package cluster

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
)

type PeerManager struct {
    Self  string
    Peers []string
}

func NewPeerManager(self string, peers []string) *PeerManager {
    filtered := []string{}
    for _, p := range peers {
        if p != "" && p != self {
            filtered = append(filtered, p)
        }
    }
    return &PeerManager{Self: self, Peers: filtered}
}

func (pm *PeerManager) Replicate(key, value string) {
    for _, peer := range pm.Peers {
        go func(p string) {
            data := map[string]string{"key": key, "value": value}
            jsonData, _ := json.Marshal(data)
            _, err := http.Post(p+"/replicate", "application/json", bytes.NewBuffer(jsonData))
            if err != nil {
                log.Printf("Replication to %s failed: %v", p, err)
            }
        }(peer)
    }
}