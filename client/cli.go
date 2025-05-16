package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

func main() {
    if len(os.Args) < 4 {
        fmt.Println("Usage: go run cli.go <set|get> <key> [value]")
        return
    }

    cmd := os.Args[1]
    key := os.Args[2]
    node := "http://localhost:8080"

    switch cmd {
    case "set":
        if len(os.Args) < 4 {
            fmt.Println("set requires a key and value")
            return
        }
        value := os.Args[3]
        data := map[string]string{"key": key, "value": value}
        b, _ := json.Marshal(data)
        http.Post(node+"/set", "application/json", bytes.NewBuffer(b))
        fmt.Println("Set successful")

    case "get":
        res, err := http.Get(node + "/get/" + key)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        defer res.Body.Close()
        var data map[string]string
        json.NewDecoder(res.Body).Decode(&data)
        fmt.Println("Value:", data["value"])
    }
}