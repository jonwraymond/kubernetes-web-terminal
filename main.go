package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		return true
	},
}

type TerminalSession struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

type Pod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
}

type ScriptRequest struct {
	Script string `json:"script"`
	Type   string `json:"type"`
}

func main() {
	router := mux.NewRouter()

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API endpoints
	router.HandleFunc("/api/pods", getPodsHandler).Methods("GET")
	router.HandleFunc("/api/terminal", terminalHandler).Methods("GET")
	router.HandleFunc("/api/execute-script", executeScriptHandler).Methods("POST")

	// Serve index.html for root path
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func getKubeClient() (*kubernetes.Clientset, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(config)
}

func getPodsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := getKubeClient()
	if err != nil {
		log.Printf("Failed to get Kubernetes client: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pods": [], "error": "Failed to connect to Kubernetes cluster"}`))
		return
	}

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pods": [], "error": "Failed to list pods"}`))
		return
	}

	var podList []Pod
	for _, pod := range pods.Items {
		podList = append(podList, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pods": podList,
	})
}

func terminalHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket terminal handler
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Terminal handling logic will go here
}

// Next implements remotecommand.TerminalSizeQueue
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	size := <-t.sizeChan
	return &size
}

// Read implements io.Reader
func (t *TerminalSession) Read(p []byte) (int, error) {
	_, message, err := t.wsConn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(p, message)
	return len(message), nil
}

func executeScriptHandler(w http.ResponseWriter, r *http.Request) {
	var req ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For demonstration purposes, we'll simulate script execution
	// In a real implementation, you would execute the script in a secure environment

	w.Header().Set("Content-Type", "text/plain")

	switch req.Type {
	case "bash":
		fmt.Fprintf(w, "Executing bash script:\n%s\n\n--- Simulated Output ---\nScript executed successfully!\nNote: In production, this would run in a controlled Kubernetes environment.", req.Script)
	case "python":
		fmt.Fprintf(w, "Executing python script:\n%s\n\n--- Simulated Output ---\nPython script executed successfully!\nNote: In production, this would run in a controlled Kubernetes environment.", req.Script)
	case "kubectl":
		fmt.Fprintf(w, "Executing kubectl commands:\n%s\n\n--- Simulated Output ---\nkubectl commands executed successfully!\nNote: In production, this would run actual kubectl commands against the cluster.", req.Script)
	default:
		fmt.Fprintf(w, "Executing %s script:\n%s\n\n--- Simulated Output ---\nScript executed successfully!", req.Type, req.Script)
	}
}

// Write implements io.Writer
func (t *TerminalSession) Write(p []byte) (int, error) {
	err := t.wsConn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
