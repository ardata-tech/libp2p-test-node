# Basic Libp2p Test Node

This is a basic libp2p test node that can be used to test the libp2p network. It is a simple node that can be used to connect to the network and send messages to other nodes.

## Build
```
make build
go build -o libp2p-node
```
## Running

To run the node, you can use the following command:

```bash
./libp2p-node --listen-port=4001
./libp2p-node --listen-addr=4002 --bootstrap-peers=/ip4/<ip>/tcp/4001/p2p/<peer-id>,/ip4/<ip>/tcp/4001/p2p/<peer-id>

2024/07/09 20:50:32 Node ID: QmPeVM6s8GxxSoujsrfwGLHfgmN5igxKr7Aqvv6YA5tN3d
2024/07/09 20:50:32 Listening on: /ip4/0.0.0.0/tcp/4002
```

## Testing

To test the peering, verifying and signing, you can do the following

```bash
./libp2p-node --listen-port=4001

// on another terminal
./libp2p-node --listen-port=4002

// on another terminal
./libp2p-node --listen-port=4003

// on another terminal
./libp2p-node --listen-port=4004
``` 

## API 
The node exposes an endpoint to check the eth price and connected nodes
```
curl http://localhost:8080/get-eth-prices
curl http://localhost:8080/get-connected-peers (WIP)
```

## Docker
To run the node in a docker container, you can use the following command:

```bash
docker build -t libp2p-node .
docker run -p 8080:8080 -p 4001:4001 libp2p-node -e LISTEN_PORT=4001 libp2p-node
```

## Notes
- The node adds a postgres and sqlite DB so it supports both
- The node has a basic REST API to check the eth price. Connected peers is WIP.

