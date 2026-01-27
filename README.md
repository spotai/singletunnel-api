# singletunnel-api

Public declarations of Single Tunnel Communications project.

This repository contains the Protocol Buffer definitions and Go-generated code for the gRPC services used by the SingleTunnel system.

## Documentation

For a comprehensive overview of the system architecture, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Contents

- `coord/coord.proto` - Protocol Buffer definitions for coordination services
- `coord/coord.pb.go` - Generated Go code for protobuf messages
- `coord/coord_grpc.pb.go` - Generated Go code for gRPC services
- `clntagent/config.go` - Client agent configuration structures

## Services

### ForServers
- `GetServerConfig` - VPN servers poll for configuration updates
- `ShuttingDown` - Servers notify coordination server on shutdown

### ForClients
- `ListServers` - Query available VPN gateway servers
- `Connect` - Establish connection to a specific server
- `Disconnect` - Gracefully disconnect from current server
- `WaitEvent` - Long-poll for state change notifications
