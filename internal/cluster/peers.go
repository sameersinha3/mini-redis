package cluster

import (
    "bytes"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "sort"
    "fmt"
    "sync"
)

const (
    replicationTimeout = 5 * time.Second
    maxRetries         = 3
    retryDelay         = 1 * time.Second
)

type PeerManager struct {
    Self      string
    Peers     []string
    LivePeers []string
    Leader    string
    mu        sync.RWMutex
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

func (pm *PeerManager) ElectLeader() {
    all := append(pm.Peers, pm.Self)
    sort.Strings(all) 
    pm.Leader = all[0]
}

func (pm *PeerManager) ElectLeaderFromLive() {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    livePeers := []string{}
    for _, peer := range pm.LivePeers {
        livePeers = append(livePeers, peer)
    }
    livePeers = append(livePeers, pm.Self)

    sort.Strings(livePeers)
    pm.Leader = livePeers[0]
    fmt.Printf("New leader elected from live peers: %s\n", pm.Leader)
}


func (pm *PeerManager) IsLeader() bool {
    return pm.Self == pm.Leader
}


func (pm *PeerManager) UpdateLivePeers() {
    var live []string
    for _, peer := range pm.Peers {
        if peer == pm.Self {
            live = append(live, peer)
            continue
        }

        url := fmt.Sprintf("http://%s/leader", peer)
        client := http.Client{Timeout: 1 * time.Second}
        resp, err := client.Get(url)
        if err == nil && resp.StatusCode == http.StatusOK {
            live = append(live, peer)
        }
        if resp != nil {
            resp.Body.Close()
        }
    }

    pm.mu.Lock()
    pm.LivePeers = live
    pm.mu.Unlock()
}

func (pm *PeerManager) IsLeaderAlive() bool {
    if pm.Leader == "" {
        return false
    }

    url := fmt.Sprintf("http://%s/leader", pm.Leader)
    client := http.Client{Timeout: 1 * time.Second}

    resp, err := client.Get(url)
    if err != nil {
        fmt.Printf("Failed to reach leader %s: %v\n", pm.Leader, err)
        return false
    }
    defer resp.Body.Close()

    return resp.StatusCode == http.StatusOK
}


func (pm *PeerManager) StartLeaderMonitor() {
    go func() {
        failCount := 0
        ticker := time.NewTicker(3 * time.Second)

        for range ticker.C {
            pm.UpdateLivePeers()

            if pm.IsLeader() {
                continue
            }

            if pm.IsLeaderAlive() {
                failCount = 0
            } else {
                failCount++
                if failCount >= 3 {
                    fmt.Println("Leader is down! Re-electing from live peers...")
                    pm.ElectLeaderFromLive()
                    failCount = 0
                }
            }
        }
    }()
}

