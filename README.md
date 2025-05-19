# Distributed KV Store in Go

A distributed key-value store implemented in Go, using Raft consensus for leader election, replication, and snapshotting.

## Features
* Raft-based leader election and log replication
* Snapshotting for state persistence and recovery
* Basic CRUD operations via REST API (/api/set, /api/get/{key}, /api/delete/{key})
* Multi-node cluster support

## Usage

### Start 2 nodes in separate terminals:
```bash
PORT=8080 PEERS=localhost:8081 go run cmd/kvnode/main.go
PORT=8081 PEERS=localhost:8080 go run cmd/kvnode/main.go
```

### API Endpoints

#### Check leader
```bash
curl http://localhost:8080/api/leader
```

#### Set Key
```bash
curl -X POST http://localhost:8080/api/set \
     -H "Content-Type: application/json" \
     -d '{"key":"foo","value":"bar"}'
```

#### Get Key
```bash
curl http://localhost:8080/api/get/foo
```

#### Delete Key
```bash
curl -X DELETE http://localhost:8080/api/delete/foo
```

## Snapshotting
The Raft leader automatically takes snapshots of the store state every 10 seconds to persist progress and help with faster recovery.

## Docker
We can build and run with Docker
```bash
docker build -t kvnode .

docker run -e PORT=8080 -e PEERS=host.docker.internal:8081 -p 8080:8080 kvnode
docker run -e PORT=8081 -e PEERS=host.docker.internal:8080 -p 8081:8081 kvnode
```

## Important Notes
* Write operations (set, delete) must be sent to the current leader; followers will reject them.
* Read operations (get) can be handled by any node.
* Allow a few seconds after startup for leader election to complete.