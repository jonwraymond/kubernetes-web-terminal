# FileMount Integration with TerminalConfig CRD

This document describes the FileMount integration feature added to the Kubernetes Web Terminal project.

## Overview

The TerminalConfig Custom Resource Definition (CRD) now supports file mount references, allowing you to mount ConfigMaps, Secrets, and Volumes into terminal sessions. This enables users to access configuration files, secrets, and persistent data directly from their web terminal sessions.

## TerminalConfig CRD

The `TerminalConfig` CRD defines the configuration for a terminal session with the following key features:

### Spec Fields

- **image**: Container image to use for the terminal (default: ubuntu:22.04)
- **command**: Command to run in the terminal (default: ["/bin/bash"])
- **args**: Arguments to pass to the command
- **fileMounts**: Array of file mount definitions
- **resources**: Resource requirements for the terminal container
- **securityContext**: Security context for the terminal container

### FileMount Types

The `fileMounts` field supports three types of references:

#### 1. ConfigMap Reference
```yaml
fileMounts:
- name: config-files
  mountPath: /etc/config
  configMapRef:
    name: app-config
  readOnly: true
```

#### 2. Secret Reference
```yaml
fileMounts:
- name: secret-files
  mountPath: /etc/secrets
  secretRef:
    secretName: app-secrets
  readOnly: true
```

#### 3. Volume Reference
```yaml
fileMounts:
- name: data-volume
  mountPath: /data
  volumeRef:
    name: shared-data
    subPath: app-data
  readOnly: false
```

## API Endpoints

The following new API endpoints are available:

- `GET /api/terminalconfigs` - List all TerminalConfigs
- `GET /api/terminalconfigs/{name}` - Get a specific TerminalConfig
- `POST /api/terminalconfigs` - Create a new TerminalConfig
- `GET /api/terminal?config={name}` - Connect to a terminal using a specific TerminalConfig

## Installation

1. Install the CRD:
```bash
kubectl apply -f manifests/terminalconfig-crd.yaml
```

2. Create example resources (optional):
```bash
kubectl apply -f examples/example-resources.yaml
```

3. Create a TerminalConfig:
```bash
kubectl apply -f examples/example-terminalconfig.yaml
```

## Usage Example

1. Create a TerminalConfig with file mounts:
```bash
kubectl apply -f examples/example-terminalconfig.yaml
```

2. Connect to the terminal via the web interface:
```
http://localhost:8080/api/terminal?config=example-terminal
```

3. The terminal session will start with the specified file mounts available at their configured paths.

## Status and Conditions

The TerminalConfig status includes:

- **phase**: Current phase (Pending, Running, Failed, Terminated)
- **message**: Additional information about the current phase
- **conditions**: Array of condition objects indicating readiness and file mount status
- **createdAt**: Timestamp when the terminal session was created

### Condition Types

- **Ready**: Indicates whether the terminal config is ready
- **FilesMounted**: Indicates whether the file mounts are successfully mounted

## Security Considerations

- File mounts are subject to Kubernetes RBAC policies
- The security context can be configured to run as non-root
- ReadOnly mounts are recommended for configuration and secret files
- Volume references should point to appropriate storage backends

## Implementation Details

The integration consists of:

1. **API Types** (`pkg/apis/terminal/v1/types.go`): Go structs defining the TerminalConfig CRD
2. **CRD Manifest** (`manifests/terminalconfig-crd.yaml`): Kubernetes CRD definition
3. **Client** (`pkg/client/terminalconfig.go`): Client for interacting with TerminalConfig resources
4. **Server Integration** (`main.go`): Updated server with TerminalConfig support

## Future Enhancements

Potential future improvements include:

- Pod controller to actually create pods with mounted volumes
- WebSocket integration with real Kubernetes exec sessions  
- File mount validation and status reporting
- Support for additional volume types (EmptyDir, HostPath, etc.)
- Terminal session management and cleanup