# Kubernetes Web Terminal

A web-based terminal interface for Kubernetes clusters built with Go and WebSockets.

## Features

- Web-based terminal access to Kubernetes pods
- Real-time terminal interaction via WebSockets
- Pod listing and selection
- Support for both in-cluster and kubeconfig authentication

## Prerequisites

- Go 1.21 or higher
- Access to a Kubernetes cluster
- kubectl configured (for local development)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/jraymond/kubernetes-web-terminal.git
cd kubernetes-web-terminal
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

The server will start on port 8080 by default. You can change this by setting the `PORT` environment variable.

## Usage

1. Open your browser and navigate to `http://localhost:8080`
2. Select a pod from the list
3. Click "Connect" to open a terminal session

## Development

This project uses:
- Gorilla WebSocket for real-time communication
- Gorilla Mux for HTTP routing
- Kubernetes client-go for cluster interaction

## License

MIT License 