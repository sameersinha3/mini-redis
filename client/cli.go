package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run cli.go <set|get|delete> <key> [value]")
        return
    }

    cmd := os.Args[1]
    key := os.Args[2]
    node := getNodeAddress()

    switch cmd {
    case "set":
        if len(os.Args) < 4 {
            fmt.Println("set requires a key and value")
            return
        }
        value := os.Args[3]
        setKey(node, key, value)
    
    case "get":
        getKey(node, key)
    
    case "delete":
        deleteKey(node, key)
    
    default:
        fmt.Println("Unknown command. Use set, get, or delete")
    }
}

func getNodeAddress() string {
    node := os.Getenv("NODE")
    if node == "" {
        node = "http://localhost:8080"
    }
    return node
}

func setKey(node, key, value string) {
    data := map[string]string{"key": key, "value": value}
    b, _ := json.Marshal(data)
    
    resp, err := http.Post(node+"/set", "application/json", bytes.NewBuffer(b))
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        fmt.Println("Set successful")
    } else {
        fmt.Printf("Set failed with status code: %d\n", resp.StatusCode)
    }
}

func getKey(node, key string) {
    res, err := http.Get(node + "/get/" + key)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer res.Body.Close()
    
    if res.StatusCode == http.StatusNotFound {
        fmt.Printf("Key '%s' not found\n", key)
        return
    }
    
    var data map[string]string
    json.NewDecoder(res.Body).Decode(&data)
    fmt.Println("Value:", data["value"])
}

func deleteKey(node, key string) {
    client := &http.Client{}
    req, err := http.NewRequest("DELETE", node+"/delete/"+key, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }
    
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        fmt.Println("Delete successful")
    } else if resp.StatusCode == http.StatusNotFound {
        fmt.Printf("Key '%s' not found\n", key)
    } else {
        fmt.Printf("Delete failed with status code: %d\n", resp.StatusCode)
    }
}