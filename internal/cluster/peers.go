package cluster

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"
)

const (
    replicationTimeout = 5 * time.Second
    maxRetries         = 3
    retryDelay         = 1 * time.Second
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
            
            pm.sendWithRetry(p+"/replicate", jsonData)
        }(peer)
    }
}

func (pm *PeerManager) ReplicateDelete(key string) {
    for _, peer := range pm.Peers {
        go func(p string) {
            data := map[string]string{"key": key}
            jsonData, _ := json.Marshal(data)
            
            pm.sendWithRetry(p+"/replicate/delete", jsonData)
        }(peer)
    }
}

func (pm *PeerManager) sendWithRetry(url string, jsonData []byte) {
    client := &http.Client{
        Timeout: replicationTimeout,
    }

    var err error
    for retry := 0; retry < maxRetries; retry++ {
        req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        
        resp, reqErr := client.Do(req)
        if reqErr == nil {
            resp.Body.Close()
            if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                return // Success
            }
            err = reqErr
        } else {
            err = reqErr
        }
        
        log.Printf("Replication to %s failed (attempt %d/%d): %v", 
            url, retry+1, maxRetries, err)
        
        // Wait before retrying
        time.Sleep(retryDelay)
    }
    
    log.Printf("Replication to %s permanently failed after %d attempts", 
        url, maxRetries)
}