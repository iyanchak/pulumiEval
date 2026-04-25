# Refactoring Plan

## 1. Import Paths That Will Change
- **Root `main.go`:** Will be updated to import the newly extracted core business logic package, e.g., `pulumiEval/pkg` (or `pulumiEval/pkg/infra`).
- **`counter/go.mod`:** The module path must be changed from `pulumiEval` to `counter` (or `pulumiEval/counter`) to support Go Workspaces (`go.work`) without causing duplicate module naming conflicts with the root `go.mod`.

## 2. Pack Build Command
To replace the multi-stage Dockerfile with Cloud Native Buildpacks, we will use the following `pack` CLI command with a Google-native builder:

```bash
pack build localhost:32000/counter-server:latest --builder gcr.io/buildpacks/builder:v1 --path ./counter
```
*(Note: If strict CGO disabling is required as it was in the Dockerfile (`CGO_ENABLED=0`), we can append `--env CGO_ENABLED=0` to the command.)*

## 3. Implementation Steps Overview
1.  **Extract Pulumi Logic:** Create a `pkg` directory (e.g., `/pkg/infra`) and move the `createResources` function from the root `main.go` into it.
2.  **Setup Go Workspace:** Create a `go.work` file at the root containing `use ( . ./counter )` and rename the module in `counter/go.mod`.
3.  **Replace Dockerfile:** Delete `counter/Dockerfile`. Replace the `docker.NewImage` resource in Pulumi with a local command execution (e.g., using `pulumi-command`) that runs the `pack build` command, or build the image externally.
4.  **Update Root App:** Update the root `main.go` to import and invoke the extracted logic from the new package.

append `--env CGO_ENABLED=0` and
"We are refactoring this project into a Go Monorepo and switching to Cloud Native Buildpacks.

Context: > - Deployment target: Local microk8s.

Build Tool: pack CLI (Buildpacks).

Registry: localhost:32000 (local microk8s registry).

Phase 1: Monorepo Setup (Do this first)

Use your tools to read the current project.

Create a /pkg directory for shared logic and a /cmd directory for the app.

Initialize a Go Workspace (go.work) at the root.

Update the imports so the Pulumi code and the Go app use the new monorepo paths.

Phase 2: Pulumi 'Pack' Integration

Replace the @pulumi/docker Image resource with @pulumi/command.

The command should run: pack build localhost:32000/[app-name]:latest --builder gcr.io/buildpacks/builder:google-22 --publish.

Update the Kubernetes Deployment to use that image."
