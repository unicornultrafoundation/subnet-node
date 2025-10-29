## K8s package

Kubernetes integration for Subnet Node. This package manages the full lifecycle of deployments on a Kubernetes cluster: manifests → workloads → monitoring → expiry/teardown, with service-aware scaling to zero and back.

### Highlights
- Service orchestrator with internal state machine (deploymentManager)
- Kubernetes client with builders for Deployments/StatefulSets/Services/NetworkPolicies
- Health monitoring and expiry handling
- Scale-to-zero and scale-back per service

### Architecture
- `service.go` — Top-level orchestrator. Owns deployment managers, routes bus events, exposes methods used by HTTP handlers. Entry points for scale-to-zero/back and status APIs.
- `manager.go` — deploymentManager per Lease. Manages state transitions, deployment updates, scaling state machine, and starts/stops monitoring.
- `monitor.go` — deploymentMonitor that checks health/availability; respects scaled-down state.
- `kube/client.go` — Concrete Kubernetes client: CRUD, patching replicas, pod/log exec, namespace management, CRDs access. Wraps client-go calls with consistent error mapping.
- `kube/builder/*` — Builders for Kubernetes objects derived from manifest: Deployment, StatefulSet, Service, NetPolicy, etc.
- `expiry.go` — Lease expiry integration and actions (scale down or delete).
- `types/v1/*` — Public types/interfaces for deployment description, queries, provider client, inventory client, etc.
- `mocks/*` — Test doubles for unit tests.

### IDs and Namespacing
All workloads are isolated per Lease using a namespace derived from `LeaseID` (owner, dseq, gseq, oseq, provider). See `util/lease_id_to_namespace.go` and `kube/builder/settings.go`.

### Scaling
- Methods on Service:
  - `ScaleToZero(ctx, leaseID)` — Batch scales all services in a lease to 0 replicas.
  - `ScaleBack(ctx, leaseID)` — Restores replicas to original manifest counts.
- Client method:
  - `ScaleServices(ctx, leaseID, serviceReplicas map[string]int32)` — Scales Deployments/StatefulSets for the given service names only. Uses JSON Merge Patch (`{"spec":{"replicas":N}}`).
- Manager scaling state machine (`manager.go`):
  - States: active → scaling-down → scaled-down → scaling-up → active
  - Safe concurrent transitions, cancellation-aware
  - Stores original replicas from current manifest on first scale-down
  - Monitor considers scaled-down leases as healthy

### HTTP API (mounted under /k8s by core/corehttp)
Defined in `internal/api/k8s.go`; served via chi router.
- `GET  /api/v1/status` — Cluster status
- `GET  /api/v1/leases` — All leases summary
- `GET  /api/v1/leases/{owner}/{dseq}` — Lease status
- `GET  /api/v1/leases/{owner}/{dseq}/manifest` — Manifest for lease
- `POST /api/v1/deployments` — Create/request deployment
- WebSocket
  - `GET /api/v1/deployments/{owner}/{dseq}/ws/exec`
  - `GET /api/v1/deployments/{owner}/{dseq}/ws/logs`

Routing tip: with `net/http.ServeMux`, mount the `/k8s/` subtree and strip the `/k8s` prefix so the inner router sees paths starting with `/` (e.g. `/api/v1/status`):
```go
mux.Handle("/k8s/", http.StripPrefix("/k8s", k8sHandler.Router()))
```

### Inventory
`kube/operators/clients/inventory` provides capacity and scheduling info; used to validate and plan replicas.

### Configuration
Read via repo config in `core/node/k8s.go`:
- `deployer.enable` (bool): enable k8s deployer
- `deployer.kubeconfig_path` (string): kubeconfig path (required)
- `vpn.virtual_ip` (string): provide the ip that users can access the deployed services (now using VPN)

### Development
- Tests: run package tests
```bash
go test ./core/k8s/... -count=1
```
- Integration test requires a reachable kube cluster and kubeconfig (see `kube/k8s_integration_test.go`).

### Key Files
- `client.go` — Service interface + null client
- `kube/client.go` — Kubernetes implementation
- `service.go` — Orchestrator, public surface
- `manager.go` — Per-lease lifecycle & scaling state
- `monitor.go` — Health checks
- `expiry.go` — Lease expiry actions
- `manifest/*` — Manifest parsing/validation helpers
- `types/v1/*` — Public types/interfaces

### Notes
- Addresses in LeaseIDs are normalized to lowercase at runtime and in tests.
- Scaling patches use MergePatch; do not use JSONPatch with object bodies.


