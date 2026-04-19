# Pulumi Go – Kubernetes Counter Service
Provision a stateful counter service on Kubernetes with Pulumi and Go.

## Description
This repository contains a Pulumi program written in Go that provisions a Kubernetes workload consisting of a PersistentVolumeClaim, a Deployment, and a Service for a simple counter micro-service. It leverages Infrastructure as Code (IaC) to ensure reproducible deployments, utilizing stateful storage for the counter data, node affinity to target worker nodes, and automatic NodePort allocation for external access.

## Architecture
```mermaid
graph LR
    PVC[sqlite-pvc] --> Deployment[counter-deployment]
    Deployment --> Service[counter-service]
    Service --> NodePort[NodePort (auto)]
    subgraph Worker Nodes
        node1[worker-1]
        node2[worker-2]
    end
    Service -->|affinity| Worker Nodes
```

### Components
- **PersistentVolumeClaim (`sqlite-pvc`)**: 1Gi storage using the `microk8s-hostpath` storage class for SQLite persistence.
- **Deployment (`counter-deployment`)**: Runs `ghcr.io/sidpalas/devops-directive-kubernetes-course/counter-service:v1.0.0` with node affinity for `node-role.kubernetes.io/worker=true`.
- **Service (`counter-service`)**: A `NodePort` service exposing the application on port 8080.

## Prerequisites
- Go 1.22+
- Pulumi CLI v3.x
- Pulumi Kubernetes SDK v4.x
- Access to a Kubernetes cluster (e.g., MicroK8s)
- Docker (for pulling the container image)
- `kubectl` (optional, for manual inspection)

## Getting Started
1. **Clone the repo**
   ```bash
   git clone <repository-url>
   cd pulumiEval
   ```
2. **Install dependencies**
   ```bash
   go mod tidy
   ```
3. **Login to Pulumi**
   ```bash
   pulumi login
   ```
4. **Create/Select a stack**
   ```bash
   pulumi stack init dev
   ```
5. **Deploy**
   ```bash
   pulumi up
   ```
6. **Verify**
   ```bash
   kubectl get pvc,pod,svc
   ```

## Usage
Once deployed, find the allocated NodePort:
```bash
pulumi stack output nodePort
```
Interact with the service:
```bash
curl http://<node-ip>:<nodePort>/increment
```

## Testing
The project uses Pulumi mocks to validate resource creation without requiring a live cluster.
```bash
go test ./...
```
The tests verify the instantiation of the PVC, Deployment, and Service.

## Project Structure
```
├─ go.mod
├─ go.sum
├─ main.go          # Pulumi program entry point
└─ main_test.go     # Pulumi mocks & unit tests
```

## Configuration
- `kubernetes:context`: Used to select the target kubeconfig context.

## Clean-up
To remove all resources created by the stack:
```bash
pulumi destroy
```

## Contributing
1. Fork the repository.
2. Create a feature branch.
3. Run `go test ./...` to ensure no regressions.
4. Submit a Pull Request.

## License
MIT License

## References
- [Pulumi Go SDK](https://www.pulumi.com/docs/languages-go/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
