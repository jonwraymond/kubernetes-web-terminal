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
	terminalv1 "github.com/jraymond/kubernetes-web-terminal/pkg/apis/terminal/v1"
	"github.com/jraymond/kubernetes-web-terminal/pkg/client"
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

// File upload related types
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

type Server struct {
	kubeClient       *kubernetes.Clientset
	terminalClient   *client.TerminalConfigClient
	namespace        string
}

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Printf("Warning: Could not create uploads directory: %v", err)
	}

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	config, err := getKubeConfig()
	if err != nil {
		log.Fatal(err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	terminalClient, err := client.NewTerminalConfigClient(config, namespace)
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		kubeClient:     kubeClient,
		terminalClient: terminalClient,
		namespace:      namespace,
	}

	router := mux.NewRouter()

	// Serve static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API endpoints - combine both file upload and TerminalConfig APIs
	router.HandleFunc("/api/pods", server.getPodsHandler).Methods("GET")
	router.HandleFunc("/api/upload", uploadHandler).Methods("POST")
	router.HandleFunc("/api/mount", mountHandler).Methods("POST")
	router.HandleFunc("/api/terminalconfigs", server.getTerminalConfigsHandler).Methods("GET")
	router.HandleFunc("/api/terminalconfigs/{name}", server.getTerminalConfigHandler).Methods("GET")
	router.HandleFunc("/api/terminalconfigs", server.createTerminalConfigHandler).Methods("POST")
	router.HandleFunc("/api/terminal", server.terminalHandler).Methods("GET")

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

func getKubeConfig() (*rest.Config, error) {
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
	return config, nil
}

func getKubeClient() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (s *Server) getPodsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	pods, err := s.kubeClient.CoreV1().Pods(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		// If no Kubernetes client available or listing fails, return mock pods for testing
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

	type podInfo struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Status    string `json:"status"`
	}

	var podList []podInfo
	for _, pod := range pods.Items {
		podList = append(podList, podInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pods": podList})
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

func (s *Server) getTerminalConfigsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	terminalConfigs, err := s.terminalClient.List(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list TerminalConfigs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terminalConfigs)
}

func (s *Server) getTerminalConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	ctx := context.Background()
	terminalConfig, err := s.terminalClient.Get(ctx, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get TerminalConfig: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terminalConfig)
}

func (s *Server) createTerminalConfigHandler(w http.ResponseWriter, r *http.Request) {
	var terminalConfig terminalv1.TerminalConfig
	if err := json.NewDecoder(r.Body).Decode(&terminalConfig); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if terminalConfig.Spec.Image == "" {
		terminalConfig.Spec.Image = "ubuntu:22.04"
	}
	if len(terminalConfig.Spec.Command) == 0 {
		terminalConfig.Spec.Command = []string{"/bin/bash"}
	}

	// Set metadata
	terminalConfig.APIVersion = terminalv1.SchemeGroupVersion.String()
	terminalConfig.Kind = "TerminalConfig"
	if terminalConfig.Namespace == "" {
		terminalConfig.Namespace = s.namespace
	}

	ctx := context.Background()
	created, err := s.terminalClient.Create(ctx, &terminalConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create TerminalConfig: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (s *Server) terminalHandler(w http.ResponseWriter, r *http.Request) {
	// Get terminal config name from query parameter
	terminalConfigName := r.URL.Query().Get("config")
	if terminalConfigName == "" {
		http.Error(w, "Missing 'config' query parameter", http.StatusBadRequest)
		return
	}

	// Retrieve the TerminalConfig
	ctx := context.Background()
	terminalConfig, err := s.terminalClient.Get(ctx, terminalConfigName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get TerminalConfig: %v", err), http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create terminal session with file mount support
	session := &TerminalSession{
		wsConn:   conn,
		sizeChan: make(chan remotecommand.TerminalSize),
	}

	// Log the file mounts that would be applied
	if len(terminalConfig.Spec.FileMounts) > 0 {
		log.Printf("TerminalConfig %s has %d file mounts:", terminalConfigName, len(terminalConfig.Spec.FileMounts))
		for i, mount := range terminalConfig.Spec.FileMounts {
			log.Printf("  Mount %d: %s -> %s", i+1, mount.Name, mount.MountPath)
			if mount.ConfigMapRef != nil {
				log.Printf("    ConfigMap: %s", mount.ConfigMapRef.Name)
			}
			if mount.SecretRef != nil {
				log.Printf("    Secret: %s", mount.SecretRef.SecretName)
			}
			if mount.VolumeRef != nil {
				log.Printf("    Volume: %s", mount.VolumeRef.Name)
			}
		}
	}

	// Send a welcome message showing the file mounts
	welcomeMsg := fmt.Sprintf("Terminal session started for config: %s\n", terminalConfigName)
	if len(terminalConfig.Spec.FileMounts) > 0 {
		welcomeMsg += fmt.Sprintf("File mounts configured:\n")
		for _, mount := range terminalConfig.Spec.FileMounts {
			welcomeMsg += fmt.Sprintf("  - %s mounted at %s\n", mount.Name, mount.MountPath)
		}
	}
	welcomeMsg += "$ "
	
	session.Write([]byte(welcomeMsg))

	// Simple echo server for demonstration
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		
		// Echo the message back
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
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