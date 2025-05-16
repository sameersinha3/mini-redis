# Distributed KV Store in Go

## Usage

### Start 2 nodes in separate terminals:
```bash
PORT=8080 PEERS=http://localhost:8081 go run cmd/kvnode/main.go
PORT=8081 PEERS=http://localhost:8080 go run cmd/kvnode/main.go
```

### Set and Get Key
```bash
curl -X POST -H "Content-Type: application/json" -d '{"key":"foo","value":"bar"}' http://localhost:8080/set
curl http://localhost:8081/get/foo
```

## Docker

```bash
docker build -t kvnode .
docker run -e PORT=8080 -e PEERS=http://host.docker.internal:8081 -p 8080:8080 kvnode
```