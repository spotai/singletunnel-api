# SingleTunnel System Architecture

## Overview

The SingleTunnel system provides secure VPN connectivity and ingress tunneling for distributed appliances/clients to communicate with a central infrastructure. It consists of three main components coordinated through a central server:

1. **Coordination Server (CS)** - Central control plane managing VPN connections and configurations
2. **Server Agents** - Running on VPN gateway servers that accept client connections
3. **Client Agents** - Running on appliances/devices that need connectivity

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Coordination Server (CS)                      │
│  - gRPC Services (ForServers, ForClients)                        │
│  - PostgreSQL Database (servers, clients, state)                 │
│  - IPAM (IP Address Management for VPN subnet)                   │
│  - Certificate Management (OpenVPN certs, Let's Encrypt)         │
│  - DNS Management (Cloudflare integration)                       │
└───────────────┬─────────────────────────┬───────────────────────┘
                │                         │
        ┌───────┴────────┐        ┌───────┴────────┐
        │                │        │                │
┌───────▼──────────┐   ┌─▼────────▼─────┐   ┌─────▼──────────┐
│  Server Agent    │   │ Server Agent   │   │ Client Agent   │
│  (VPN Gateway 1) │   │ (VPN Gateway 2)│   │  (Appliance)   │
│                  │   │                │   │                │
│  - OpenVPN/WG    │   │  - OpenVPN/WG  │   │  - OpenVPN/WG  │
│  - Nginx Ingress │   │  - Nginx Ingress│   │  - TLS Proxies │
│  - Systemd Mgmt  │   │  - Systemd Mgmt│   │  - Route Mgmt  │
└──────────────────┘   └────────────────┘   └────────────────┘
```

## Core Components

### 1. Coordination Server (CS)

**Location**: `singletunnel/pkg/cs/`

**Purpose**: Central control plane that orchestrates all VPN connections and manages system state.

#### Key Responsibilities:
- **Server Management**: Register VPN servers, allocate VPN IPs from subnet pool (100.64.0.0/10), maintain heartbeats
- **Client Management**: Register clients, track connection state, manage client-to-server assignments
- **Configuration Distribution**: Provide dynamic configs to servers (client lists, certs) and clients (VPN settings, ingress certs)
- **Certificate Management**: 
  - OpenVPN: CA-signed certificates via EasyRSA for servers and clients
  - Ingress: Let's Encrypt certificates for client ingress endpoints
- **DNS Management**: Create/update DNS A records (Cloudflare) for:
  - Server egress endpoints (e.g., `server-id.domain.com`)
  - Client ingress endpoints (e.g., `client-id.ingress.domain.com`)
- **IP Address Management**: Allocate/release IPs from VPN subnet using go-ipam library
- **Reconciliation**: Periodic cleanup of stale servers/clients based on heartbeat timeouts

#### gRPC Services:

**ForServers Service**:
```protobuf
rpc GetServerConfig(GetServerConfigRequest) returns (GetServerConfigResponse)
rpc ShuttingDown(ShutdownNotification) returns (ShutdownResponse)
```

**ForClients Service**:
```protobuf
rpc ListServers(ClientInfo) returns (ServersInfo)
rpc Connect(ConnectionSpec) returns (ClientConfig)
rpc Disconnect(DisconnectRequest) returns (DisconnectResponse)
rpc WaitEvent(WaitEventSpec) returns (Event)
```

#### Data Store:
- PostgreSQL database with migrations
- Tables: `servers`, `clients` (with foreign key relationships)
- Tracks: ServerInfo, ClientInfo, WG IPs, heartbeats, revisions, connection state

#### Configuration:
```yaml
EventCheckTimeout: 25s          # How long WaitEvent blocks before returning "None"
EventCheckPollingInterval: 3s   # Internal polling interval for state changes
ReconcileCheckInterval: 9s      # How often to run reconciliation
LivenessHeartbeatTimeout: 2m    # Server/client considered dead after this
MinimumRequestDelay: 1s         # Rate limiting protection
```

### 2. Server Agent

**Location**: `singletunnel/pkg/srvagent/`

**Purpose**: Runs on VPN gateway servers, configures VPN services (OpenVPN or WireGuard) and Nginx ingress proxy.

#### Key Responsibilities:
- **Registration & Heartbeat**: Register with CS, poll for config updates (GetServerConfig)
- **VPN Service Management**:
  - **OpenVPN**: Write config files (`/etc/openvpn/server/tun0.conf`, certs, DH params), restart systemd service
  - **WireGuard**: Generate keypair, write config, manage wg0 interface via wg-quick
- **Ingress Proxy Management**: 
  - Parse OpenVPN client IP mappings (`/var/log/openvpn/ipp.txt`) or WireGuard peer configs
  - Generate Nginx stream config with SNI-based routing (`$ssl_preread_server_name`)
  - Reload Nginx when client list changes
- **File Caching**: Avoid unnecessary service restarts by caching config content
- **Systemd Integration**: Reload/restart services via D-Bus

#### Operational Flow:
1. Connect to CS via gRPC
2. Send ServerInfo (ID, egress IP/port, WG public key if applicable)
3. Receive ServerConfig (OpenVPN/WG config, list of connected clients)
4. Write config files and restart services
5. Poll CS every 5 seconds for updates (revision-based, only get new config if changed)
6. Monitor client IPs and update Nginx routing table

#### Nginx Configuration:
- Uses stream module with `ssl_preread` for SNI routing
- Maps FQDN → backend IP:port (VPN IP of client + proxy port)
- Example: `client-123.ingress.example.com` → `100.64.1.5:8443`

### 3. Client Agent

**Location**: `singletunnel/pkg/clntagent/`

**Purpose**: Runs on appliances/devices, establishes VPN connection and provides ingress endpoints.

#### Key Responsibilities:
- **Server Selection**: Query CS for available servers, randomly choose compatible one (OpenVPN vs WireGuard)
- **VPN Connection**:
  - **OpenVPN**: Spawn `openvpn` process with CS-provided config
  - **WireGuard**: Generate keypair, spawn `wg-quick up` with CS-provided peer config
- **Route Management**: Install policy-based routing rules using netlink:
  - Create custom routing table (default: 567) with default route via VPN
  - Add IP rules with priorities (25000-25999)
  - Exclude CS server IP and configured networks from VPN
  - Honor explicit routes in main table (suppressPrefixLength=0)
- **Ingress Proxies**: Spawn TLS proxy (ncat) for each ingress endpoint
  - Listen on VPN IP with Let's Encrypt cert
  - Forward to local service (with optional TLS)
- **Event Monitoring**: Long-poll CS for state change events, reconnect on server failure
- **Config Watching**: Use fsnotify to reload when config file changes

#### Operational Flow:
1. Load config from YAML file (watched for changes)
2. ListServers() to get available VPN gateways
3. Connect() to chosen server, receive ClientConfig (VPN settings, ingress certs)
4. Spawn VPN subprocess (OpenVPN or WireGuard)
5. Wait for VPN interface to get IP address
6. Install policy routing rules
7. Spawn TLS proxy for each ingress endpoint
8. WaitEvent() loop to detect server changes, reconnect if needed

#### Configuration Example:
```yaml
Enable: true
Client:
  Id: "appliance-001"
  WgPublicKey: ""  # Set automatically if WireguardKeyDir specified
  ProxyBasePort: 10000
  Ingresses:
    - LocalPort: 8080
      IsTls: true
      TlsServerName: "localhost"
      FqdnSuffix: "ingress.example.com"
CsAddr: "cs.example.com:6750"
Tls:
  CertFile: "/path/to/client.crt"
  KeyFile: "/path/to/client.key"
  CaFile: "/path/to/ca.crt"
WireguardKeyDir: "/etc/wireguard"  # Empty string for OpenVPN mode
```

## Networking Architecture

### VPN Modes

The system supports two VPN technologies:

#### OpenVPN Mode
- **Protocol**: TCP or UDP (configurable per server)
- **Authentication**: TLS mutual authentication with CA-signed certificates
- **Key Exchange**: Managed by OpenVPN, DH params provided by CS
- **IP Assignment**: Static IPs allocated by CS, tracked in database
- **Interface**: `tun0` device
- **Server Process**: `openvpn-server@tun0.service` (systemd)
- **Client Process**: `openvpn` binary spawned directly

#### WireGuard Mode
- **Protocol**: UDP
- **Authentication**: Public key cryptography (Ed25519)
- **Key Management**: Generated locally, public key shared with CS
- **IP Assignment**: Static IPs allocated by CS, tracked in database
- **Interface**: `wg0` device
- **Configuration**: Via `wg-quick` with TableOff=true (routing managed separately)
- **Peers**: Each client is a peer of server, server is single peer of each client

### IP Address Management

- **Subnet**: 100.64.0.0/10 (CGNAT range, 4M addresses)
- **IPAM**: go-ipam library manages allocation
- **Persistence**: IPs stored in PostgreSQL, survive CS restarts
- **Lifecycle**: Allocated on first connection, released on server shutdown

### Routing Architecture (Client Side)

Clients use Linux policy-based routing for split-tunneling:

1. **Custom Routing Table** (table 567):
   - Contains single default route via VPN interface
   
2. **IP Rules** (priority 25000-25999):
   - Priority 25000: Honor specific routes in main table (suppressPrefixLength=0)
   - Priority 25001+: Exclusions (CS server IP, configured networks) use main table
   - Priority 25xxx: Default rule - invert match on VPN server IP → use VPN table
   
3. **Effect**: All traffic goes via VPN except:
   - Traffic to VPN server itself (avoid routing loop)
   - Explicitly excluded networks (from CS config)
   - More specific routes in main table

### Ingress Architecture

Clients expose services to the internet via ingress tunnels:

```
Internet → DNS (A record) → Server Nginx (SNI routing) → Client TLS Proxy → Local Service
```

1. **DNS**: CS creates `client-id.suffix` → server ingress IP
2. **TLS Connection**: Client connects to server, SNI header contains FQDN
3. **Nginx SNI Routing**: Matches FQDN → backend VPN IP:port
4. **Client Proxy**: ncat listens on VPN IP with Let's Encrypt cert, forwards to localhost
5. **Local Service**: Application running on client receives traffic

**Example Flow**:
```
https://appliance-001.ingress.example.com
  ↓ DNS resolves to 203.0.113.10 (server ingress IP)
  ↓ TLS connection with SNI=appliance-001.ingress.example.com
  ↓ Nginx on 203.0.113.10:443 sees SNI, proxies to 100.64.1.5:10080
  ↓ ncat on appliance-001 (VPN IP 100.64.1.5) terminates TLS, forwards to localhost:8080
  ↓ Application on appliance-001 receives HTTP request
```

## Security Model

### Authentication

1. **Mutual TLS**: All gRPC communication uses mTLS
   - CS validates client CN matches claimed ID
   - Prevents impersonation attacks
   
2. **VPN Authentication**:
   - **OpenVPN**: CA-signed certificates (EasyRSA)
   - **WireGuard**: Public key cryptography
   
3. **Ingress TLS**: Let's Encrypt certificates for public endpoints

### Certificate Management

1. **OpenVPN CA** (EasyRSA):
   - CS maintains CA keypair
   - Issues server certs (CN=server-id)
   - Issues client certs (CN=client-id)
   
2. **Let's Encrypt**:
   - CS manages certificate requests
   - Certificates stored in `--lebasedir`
   - Used for client ingress endpoints

3. **Control Plane mTLS**:
   - Separate CA for CS ↔ agents
   - Certificates provided at startup (`--certs-clients`, `--certs-servers`)

### Threat Model

**Protected Against**:
- Man-in-the-middle attacks (mTLS everywhere)
- Impersonation (CN validation)
- Unauthorized access to appliances (VPN auth required)
- DNS hijacking (DNSSEC optional, TLS required)

**Assumptions**:
- CS is trusted (it issues all certificates)
- Certificate private keys are protected
- PostgreSQL access is restricted

## Protocol Flows

### Server Registration

```
Server Agent                    Coordination Server
     |                                  |
     | GetServerConfig                  |
     |  - ServerInfo (id, egress IP)    |
     | ──────────────────────────────> |
     |                                  |  1. Verify CN=id
     |                                  |  2. Check if server exists
     |                                  |  3. Allocate VPN IP
     |                                  |  4. Issue OpenVPN certs OR store WG pubkey
     |                                  |  5. Create DNS A record
     |                                  |  6. Insert/update database
     |                                  |
     | <────────────────────────────── |
     |  ServerConfig                    |
     |   - OpenVpnServerConfig OR Wg    |
     |   - Clients[] (empty initially)  |
     |   - Routes[]                     |
     |                                  |
     |  5s poll loop...                 |
```

### Client Connection

```
Client Agent                    Coordination Server                Server Agent
     |                                  |                                |
     | ListServers                      |                                |
     | ──────────────────────────────> |                                |
     | <────────────────────────────── |                                |
     |  ServersInfo (live servers)      |                                |
     |                                  |                                |
     | Connect(server-id, ClientInfo)   |                                |
     | ──────────────────────────────> |                                |
     |                                  |  1. Verify CN=client-id        |
     |                                  |  2. Check server is live       |
     |                                  |  3. Allocate VPN IP            |
     |                                  |  4. Issue OpenVPN cert OR      |
     |                                  |     store WG pubkey            |
     |                                  |  5. Issue ingress certs        |
     |                                  |  6. Update client.server       |
     |                                  |  7. Increment server.revision  |
     |                                  |  8. Create DNS A records       |
     | <────────────────────────────── |                                |
     |  ClientConfig                    |                                |
     |   - OpenVpnClientConfig OR Wg    |                                |
     |   - Ingresses[] (certs)          |                                |
     |   - ExcludedRoutes[]             |                                |
     |                                  |                                |
     | (spawn VPN, proxies)             |                                |
     |                                  |                                |
     | WaitEvent loop...                |                                |
     |                                  |                                |
     |                                  |           GetServerConfig      |
     |                                  | <────────────────────────────  |
     |                                  |            (revision changed)  |
     |                                  | ──────────────────────────────>|
     |                                  |  ServerConfig                  |
     |                                  |   - Clients[] (now includes    |
     |                                  |     new client)                |
     |                                  |                                |
     |                                  |        (server updates Nginx)  |
```

### State Change / Reconnection

```
Client Agent                    Coordination Server                Server Agent
     |                                  |                                |
     | WaitEvent (long poll)            |                                |
     | ──────────────────────────────> |                                |
     |   (blocks up to 25s)             |  Polling every 3s              |
     |                                  |                                |
     |                                  |  (server fails, reconciliation |
     |                                  |   detects timeout)             |
     |                                  |  1. UnrefServerFromClients     |
     |                                  |  2. Set client.StateChanged    |
     |                                  |                                |
     | <────────────────────────────── |                                |
     |  Event(StateChanged)             |                                |
     |                                  |                                |
     | (kill VPN, proxies)              |                                |
     | ListServers...                   |                                |
     | Connect to new server...         |                                |
```

## Data Model

### PostgreSQL Schema

**servers table**:
```sql
CREATE TABLE servers (
    id TEXT PRIMARY KEY,
    egress_ip TEXT,
    egress_port INTEGER,
    egress_protocol TEXT,
    ingress_ip TEXT,
    egress_fqdn TEXT,
    wg_public_key TEXT,
    wg_ip TEXT,  -- CIDR format
    openvpn_server_config JSONB,
    last_heartbeat TIMESTAMP,
    revision BIGINT
);
```

**clients table**:
```sql
CREATE TABLE clients (
    id TEXT PRIMARY KEY,
    info JSONB,  -- ClientInfo protobuf as JSON
    wg_ip TEXT,  -- CIDR format
    server_id TEXT REFERENCES servers(id),
    state_changed BOOLEAN,
    last_heartbeat TIMESTAMP
);
```

### Key Data Structures

**ServerInfo** (protobuf):
```protobuf
message ServerInfo {
  string id = 1;
  string egressIp = 2;       // Public IP for client VPN connections
  uint32 egressPort = 3;     // VPN listen port
  string egressProtocol = 4; // "tcp" or "udp" (OpenVPN)
  string ingressIp = 5;      // Public IP for ingress traffic
  string wgPublicKey = 7;    // WireGuard public key (empty for OpenVPN)
}
```

**ClientInfo** (protobuf):
```protobuf
message ClientInfo {
  string id = 1;
  repeated IngressInfo ingresses = 2;
  uint32 proxyBasePort = 3;
  string wgPublicKey = 4;
}

message IngressInfo {
  uint32 localPort = 2;      // Port on appliance
  bool isTls = 3;            // Is backend service TLS?
  string tlsServerName = 4;  // SNI for backend
  string fqdnSuffix = 5;     // Combined with id: "id.suffix"
}
```

## Deployment Architecture

### Coordination Server
- **Binary**: `comms cs run`
- **Dependencies**: PostgreSQL, Cloudflare API access
- **Persistence**: Database + EasyRSA/Let's Encrypt cert directories
- **Scaling**: Single instance (state in database)
- **Ports**: 6750 (clients), 6751 (servers)

### Server Agent
- **Binary**: `comms server run` (command structure unclear from code, likely custom wrapper)
- **Dependencies**: OpenVPN or WireGuard, Nginx with stream module, systemd
- **Scaling**: Multiple servers for redundancy/load balancing
- **Ports**: Configured egress port (VPN), 443 (ingress)

### Client Agent
- **Binary**: `comms client run --cfg /path/to/config.yaml`
- **Dependencies**: OpenVPN or WireGuard, ncat, root privileges (routing)
- **Scaling**: One per appliance
- **Configuration**: YAML file with fsnotify-based hot reload

## Technologies & Dependencies

### Core Stack
- **Language**: Go 1.21+
- **RPC**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with pgx driver
- **Migrations**: golang-migrate

### Networking
- **VPN**: OpenVPN, WireGuard (wg-quick)
- **Routing**: netlink library for policy-based routing
- **Proxy**: ncat (part of nmap) for TLS termination
- **Ingress**: Nginx stream module with ssl_preread

### Infrastructure
- **DNS**: Cloudflare API
- **Certificates**: EasyRSA (OpenVPN CA), Let's Encrypt (ingress)
- **IPAM**: go-ipam (Redis/etcd/MongoDB backend support)
- **Process Management**: systemd via D-Bus

### Observability
- **Logging**: logrus structured logging
- **Metrics**: DataDog StatsD support in CS

## Configuration Examples

### Coordination Server
```bash
comms cs run \
  --laddr-clients 0.0.0.0:6750 \
  --laddr-servers 0.0.0.0:6751 \
  --certs-clients /path/to/clients/certs \
  --certs-servers /path/to/servers/certs \
  --lebasedir /etc/letsencrypt \
  --ezbasedir /etc/easyrsa \
  --routes /etc/comms/default_routes.yaml \
  --domain vpn.example.com \
  --db-host postgres.local \
  --db-name comms \
  --db-migrations-dir /path/to/migrations
```

### Server Agent
```go
// Code-based configuration
config := &srvagent.Config{
    CsAddr:              "cs.example.com:6751",
    InternalIngressIp:   "10.0.1.100",
    PollIntervalInSecs:  5,
    WireguardKeyDir:     "/etc/wireguard", // or "" for OpenVPN
}

serverInfo := &api.ServerInfo{
    Id:             "gateway-001",
    EgressIp:       "203.0.113.10",
    EgressPort:     51820,
    EgressProtocol: "udp",
    IngressIp:      "203.0.113.10",
}

agent := srvagent.New(serverInfo, config)
agent.Run(context.Background())
```

### Client Agent
```yaml
# /etc/comms/client.yaml
Enable: true
Client:
  Id: "appliance-001"
  ProxyBasePort: 10000
  Ingresses:
    - LocalPort: 8080
      IsTls: true
      TlsServerName: "localhost"
      FqdnSuffix: "ingress.example.com"
CsAddr: "cs.example.com:6750"
Tls:
  CertFile: "/etc/comms/client.crt"
  KeyFile: "/etc/comms/client.key"
  CaFile: "/etc/comms/ca.crt"
WireguardKeyDir: "/etc/wireguard"
```

## Operational Considerations

### Monitoring

**Health Indicators**:
- CS database connection status
- Server/client heartbeat freshness
- IPAM allocation exhaustion
- Certificate expiration (Let's Encrypt)
- VPN process status (systemd)
- Nginx reload success rate

**Metrics** (DataDog):
- Active servers/clients count
- Connection attempts (success/failure)
- Config distribution latency
- DNS update latency
- IPAM utilization

### Failure Scenarios

1. **CS Restart**:
   - Servers/clients continue working with last config
   - Reconciliation loop waits 2x heartbeat timeout before cleanup
   - No disruption to existing connections
   
2. **Server Failure**:
   - CS reconciliation detects missing heartbeat
   - Sets StateChanged flag on connected clients
   - Clients receive event, disconnect, choose new server
   
3. **Client Restart**:
   - Re-registers with CS (reuses existing VPN IP)
   - Reconnects to last server or chooses new one
   - Server receives updated config, updates Nginx
   
4. **Database Failure**:
   - CS crashes (no state in memory)
   - Agents retry connection with exponential backoff
   
5. **Network Partition**:
   - Clients lose connectivity, retry loop
   - Servers continue serving existing clients
   - CS reconciliation eventually times out clients

### Scaling Limits

- **VPN Subnet**: 4M addresses (100.64.0.0/10)
- **Single Server**: Limited by CPU/bandwidth (OpenVPN: ~200 Mbps, WireGuard: ~1 Gbps)
- **CS**: Single instance, bottleneck is database (consider read replicas)
- **IPAM**: Backend dependent (Redis scales better than memory backend)

### Debugging

**Client Issues**:
1. Check config file syntax
2. Verify TLS cert CN matches client ID
3. Check CS connectivity: `comms client list` (if implemented)
4. Inspect VPN interface: `ip addr show tun0` or `wg show`
5. Check routing rules: `ip rule list`, `ip route show table 567`
6. Check proxy logs in `/tmp/comms-client-*/proxy_*/ncat.log`

**Server Issues**:
1. Check systemd service status: `systemctl status openvpn-server@tun0` or `wg show`
2. Check Nginx config: `nginx -t`
3. Verify OpenVPN client IP file: `/var/log/openvpn/ipp.txt`
4. Check CS connectivity and revision updates in logs

**CS Issues**:
1. Check database connection
2. Verify certificate directories exist and are readable
3. Check Cloudflare API credentials
4. Review reconciliation loop errors in logs
5. Query database directly for state inspection

## Future Considerations

### Potential Enhancements

1. **High Availability**: 
   - Multi-instance CS with leader election (etcd)
   - Server affinity to reduce thrashing
   
2. **Metrics & Observability**:
   - Prometheus integration
   - Distributed tracing (OpenTelemetry)
   - Connection quality metrics (latency, packet loss)
   
3. **Advanced Routing**:
   - Per-client custom routes
   - BGP integration for enterprise networks
   - Multi-exit VPN (route by destination)
   
4. **Performance**:
   - Connection pooling for database
   - gRPC streaming for config updates (replace polling)
   - Nginx Plus for dynamic upstream updates
   
5. **Security**:
   - Mutual TLS for ingress endpoints
   - Rate limiting and DDoS protection
   - Certificate rotation automation
   
6. **Management**:
   - Web dashboard for CS
   - CLI tools for manual intervention
   - Terraform provider for infrastructure as code

### Known Limitations

1. **Single Failure Point**: CS has no HA mechanism
2. **Polling-Based**: Server config updates use polling, not push
3. **No Traffic Shaping**: No QoS or bandwidth limiting
4. **IPv4 Only**: No IPv6 support
5. **DNS Dependency**: Relies on Cloudflare, no generic DNS API
6. **Certificate Management**: Let's Encrypt rate limits may be hit with many clients

## Glossary

- **CS**: Coordination Server
- **IPAM**: IP Address Management
- **SNI**: Server Name Indication (TLS extension)
- **mTLS**: Mutual TLS (both parties authenticate)
- **CGNAT**: Carrier-Grade NAT (RFC 6598, 100.64.0.0/10)
- **PBR**: Policy-Based Routing
- **FQDN**: Fully Qualified Domain Name
- **CN**: Common Name (X.509 certificate field)
- **Ingress**: Inbound traffic from internet to appliance
- **Egress**: Outbound traffic from appliance to internet

## References

- Protocol Buffers Definition: `singletunnel-api/coord/coord.proto`
- Coordination Server Logic: `singletunnel/pkg/cs/cs.go`
- Server Agent: `singletunnel/pkg/srvagent/srvagent.go`
- Client Agent: `singletunnel/pkg/clntagent/clntagent.go`, `run.go`
- Routing Manager: `singletunnel/pkg/routes/routes.go`
- Nginx Config: `singletunnel/pkg/nginx/nginx.go`
- WireGuard Integration: `singletunnel/pkg/wireguard/`
- OpenVPN Integration: `singletunnel/pkg/openvpn/`
