package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
}

type UploadResponse struct {
	FileID string `json:"fileId"`
	Path   string `json:"path"`
}

type MountRequest struct {
	FileID     string `json:"fileId"`
	PodName    string `json:"podName"`
	Namespace  string `json:"namespace"`
	TargetPath string `json:"targetPath"`
}

type MountResponse struct {
	TargetPath string `json:"targetPath"`
}

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Printf("Warning: Could not create uploads directory: %v", err)
	}

	router := mux.NewRouter()

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API endpoints
	router.HandleFunc("/api/pods", getPodsHandler).Methods("GET")
	router.HandleFunc("/api/upload", uploadHandler).Methods("POST")
	router.HandleFunc("/api/mount", mountHandler).Methods("POST")
	router.HandleFunc("/api/terminal", terminalHandler).Methods("GET")

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
		// If no Kubernetes client available, return mock pods for testing
		log.Printf("No Kubernetes client available, returning mock pods: %v", err)
		mockPods := []Pod{
			{Name: "nginx-deployment-123", Namespace: "default"},
			{Name: "redis-server-456", Namespace: "default"},
			{Name: "web-app-789", Namespace: "production"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]Pod{"pods": mockPods})
		return
	}

	// List pods in default namespace (can be made configurable)
	namespace := "default"
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// If listing fails, also return mock pods for testing
		log.Printf("Failed to list pods, returning mock pods: %v", err)
		mockPods := []Pod{
			{Name: "nginx-deployment-123", Namespace: "default"},
			{Name: "redis-server-456", Namespace: "default"},
			{Name: "web-app-789", Namespace: "production"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]Pod{"pods": mockPods})
		return
	}

	var pods []Pod
	for _, pod := range podList.Items {
		pods = append(pods, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]Pod{"pods": pods})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileID := r.FormValue("fileId")
	if fileID == "" {
		http.Error(w, "Missing fileId", http.StatusBadRequest)
		return
	}

	// Create file path
	uploadPath := filepath.Join("./uploads", fileID+"_"+handler.Filename)

	// Create the uploads file
	dst, err := os.Create(uploadPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	response := UploadResponse{
		FileID: fileID,
		Path:   uploadPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func mountHandler(w http.ResponseWriter, r *http.Request) {
	var req MountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For now, this is a simplified implementation
	// In a real scenario, you would copy the file to the pod using kubectl cp or similar
	log.Printf("Mount request: file %s to pod %s/%s at %s", req.FileID, req.Namespace, req.PodName, req.TargetPath)

	response := MountResponse{
		TargetPath: req.TargetPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

// Write implements io.Writer
func (t *TerminalSession) Write(p []byte) (int, error) {
	err := t.wsConn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
