package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"agent/internal/pkg/agents"
	"agent/internal/pkg/database"
	"agent/internal/pkg/templates_manager"
)

// Server represents the HTTP server for the multi-agent system
type Server struct {
	coordinator     *agents.Coordinator
	router          *mux.Router
	upgrader        websocket.Upgrader
	port            string
	projectPath     string
	activeSessions  map[string]*ChatSession
	sessionMutex    sync.RWMutex
	projectDB       *database.ProjectDB
	templateManager *templates_manager.TemplateManager
}

// ChatSession represents an active chat session
type ChatSession struct {
	ID           string
	ProjectID    string
	Connection   *websocket.Conn
	Context      context.Context
	Cancel       context.CancelFunc
	CreatedAt    time.Time
	LastActivity time.Time
	Messages     []ChatMessage
	mutex        sync.RWMutex
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	AgentType string                 `json:"agent_type,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewServer creates a new HTTP server instance
func NewServer(coordinator *agents.Coordinator, port string, projectPath string) *Server {
	router := mux.NewRouter()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}

	// Initialize database
	dbPath := filepath.Join(projectPath, "projects.db")
	projectDB, err := database.NewProjectDB(dbPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
	}

	// Initialize template manager
	templatesPath := "./templates"
	templateManager := templates_manager.NewTemplateManager(templatesPath, projectPath)

	server := &Server{
		coordinator:     coordinator,
		router:          router,
		upgrader:        upgrader,
		port:            port,
		projectPath:     projectPath,
		activeSessions:  make(map[string]*ChatSession),
		sessionMutex:    sync.RWMutex{},
		projectDB:       projectDB,
		templateManager: templateManager,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Add CORS middleware
	s.router.Use(s.corsMiddleware)

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// WebSocket endpoint for streaming responses
	api.HandleFunc("/chat/stream", s.handleWebSocketChat).Methods("GET")

	// REST endpoints
	api.HandleFunc("/chat", s.handleChatRequest).Methods("POST", "OPTIONS")
	api.HandleFunc("/chat/{sessionId}", s.handleGetChatSession).Methods("GET", "OPTIONS")
	api.HandleFunc("/chat/{sessionId}/cancel", s.handleCancelRequest).Methods("POST", "OPTIONS")
	api.HandleFunc("/chat/sessions", s.handleListChatSessions).Methods("GET", "OPTIONS")
	api.HandleFunc("/status", s.handleStatusRequest).Methods("GET", "OPTIONS")
	api.HandleFunc("/agents", s.handleAgentsList).Methods("GET", "OPTIONS")
	api.HandleFunc("/models", s.handleModelsRequest).Methods("GET", "OPTIONS")

	// Project management endpoints
	api.HandleFunc("/projects", s.handleProjectsRequest).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects", s.handleCreateProject).Methods("POST", "OPTIONS")
	api.HandleFunc("/projects/{id}", s.handleGetProject).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects/{name}/preview", s.handleProjectPreview).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects/{name}/files", s.handleProjectFiles).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects/{name}/files/{filepath:.*}", s.handleFileContent).Methods("GET", "POST", "OPTIONS")

	// Template management endpoints
	api.HandleFunc("/templates", s.handleTemplatesList).Methods("GET", "OPTIONS")

	// Health check
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")

	// Static files (for web interface)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
}

// corsMiddleware handles CORS for all requests
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Message     string `json:"message"`
	SessionID   string `json:"session_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectPath string `json:"project_path,omitempty"`
}

// ChatResponse represents the outgoing chat response
type ChatResponse struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content,omitempty"`
	SessionID string                 `json:"session_id"`
	ProjectID string                 `json:"project_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	AgentType string                 `json:"agent_type,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Progress  int                    `json:"progress,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// handleChatRequest handles REST API chat requests
func (s *Server) handleChatRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	sessionID := generateSessionID()

	// Log HTTP request start
	log.Printf("[HTTP_REQUEST_START] method=%s path=%s session_id=%s timestamp=%s user_agent=%s",
		r.Method, r.URL.Path, sessionID, startTime.Format(time.RFC3339), r.UserAgent())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[HTTP_REQUEST_ERROR] session_id=%s error=invalid_json", sessionID)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		req.SessionID = sessionID
	}

	// Log chat request
	log.Printf("[CHAT_REQUEST] session_id=%s message_length=%d project=%s",
		req.SessionID, len(req.Message), req.ProjectName)

	// Process the request through the coordinator
	err := s.coordinator.ProcessUserRequest(req.Message)
	if err != nil {
		log.Printf("[CHAT_REQUEST_ERROR] session_id=%s error=%s", req.SessionID, err.Error())
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// For REST API, we return a simple acknowledgment
	response := ChatResponse{
		Type:      "acknowledgment",
		Content:   "Request received and processing started",
		SessionID: req.SessionID,
		Timestamp: time.Now(),
		Status:    "processing",
	}

	duration := time.Since(startTime)
	log.Printf("[HTTP_REQUEST_SUCCESS] session_id=%s duration_ms=%d status=%d",
		req.SessionID, duration.Milliseconds(), http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// handleWebSocketChat handles WebSocket connections for streaming responses
func (s *Server) handleWebSocketChat(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()
	startTime := time.Now()

	log.Printf("[WEBSOCKET_CONNECT] session_id=%s timestamp=%s user_agent=%s",
		sessionID, startTime.Format(time.RFC3339), r.UserAgent())

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WEBSOCKET_ERROR] session_id=%s error=upgrade_failed", sessionID)
		return
	}
	defer conn.Close()

	// Create context with cancellation for this session
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create chat session
	session := &ChatSession{
		ID:           sessionID,
		Connection:   conn,
		Context:      ctx,
		Cancel:       cancel,
		CreatedAt:    startTime,
		LastActivity: startTime,
		Messages:     make([]ChatMessage, 0),
	}

	// Store the session
	s.sessionMutex.Lock()
	s.activeSessions[sessionID] = session
	s.sessionMutex.Unlock()

	// Send connection acknowledgment
	ack := ChatResponse{
		Type:      "connection",
		Content:   "WebSocket connected",
		SessionID: sessionID,
		Timestamp: time.Now(),
		Status:    "connected",
	}

	if err := conn.WriteJSON(ack); err != nil {
		log.Printf("[WEBSOCKET_ERROR] session_id=%s error=write_failed", sessionID)
		return
	}

	// Handle incoming messages
	for {
		var req ChatRequest
		if err := conn.ReadJSON(&req); err != nil {
			log.Printf("[WEBSOCKET_DISCONNECT] session_id=%s duration_ms=%d",
				sessionID, time.Since(startTime).Milliseconds())
			break
		}

		req.SessionID = sessionID
		session.LastActivity = time.Now()

		log.Printf("[WEBSOCKET_MESSAGE] session_id=%s message_length=%d",
			sessionID, len(req.Message))

		// Add user message to session
		userMsg := ChatMessage{
			ID:        uuid.New().String(),
			Type:      "user",
			Content:   req.Message,
			Timestamp: time.Now(),
			Status:    "received",
		}
		session.mutex.Lock()
		session.Messages = append(session.Messages, userMsg)
		session.mutex.Unlock()

		// If project ID provided, use it; otherwise create new project
		projectID := req.ProjectID
		if projectID == "" {
			projectID = uuid.New().String()
			session.ProjectID = projectID
		}

		// Send processing status
		processing := ChatResponse{
			Type:      "status",
			Content:   "Processing your request...",
			SessionID: sessionID,
			ProjectID: projectID,
			Timestamp: time.Now(),
			Status:    "processing",
			Progress:  10,
		}
		conn.WriteJSON(processing)

		// Process through coordinator with context and progress updates
		go s.processRequestWithProgress(session, req, projectID)
	}

	// Clean up session when connection closes
	s.sessionMutex.Lock()
	delete(s.activeSessions, sessionID)
	s.sessionMutex.Unlock()
}

// processRequestWithProgress handles request processing with progress updates
func (s *Server) processRequestWithProgress(session *ChatSession, req ChatRequest, projectID string) {
	conn := session.Connection

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PROCESS_ERROR] session_id=%s error=panic recovered: %v", session.ID, r)
			errorResp := ChatResponse{
				Type:      "error",
				Content:   "Internal processing error",
				SessionID: session.ID,
				ProjectID: projectID,
				Timestamp: time.Now(),
				Status:    "error",
			}
			conn.WriteJSON(errorResp)
		}
	}()

	// Send progress updates
	progressSteps := []struct {
		progress int
		content  string
	}{
		{20, "Analyzing request..."},
		{40, "Creating project setup..."},
		{60, "Generating components..."},
		{80, "Finalizing code..."},
		{95, "Almost done..."},
	}

	for _, step := range progressSteps {
		select {
		case <-session.Context.Done():
			// Request was cancelled
			cancelled := ChatResponse{
				Type:      "cancelled",
				Content:   "Request was cancelled",
				SessionID: session.ID,
				ProjectID: projectID,
				Timestamp: time.Now(),
				Status:    "cancelled",
			}
			conn.WriteJSON(cancelled)
			return
		default:
			progress := ChatResponse{
				Type:      "progress",
				Content:   step.content,
				SessionID: session.ID,
				ProjectID: projectID,
				Timestamp: time.Now(),
				Status:    "processing",
				Progress:  step.progress,
			}
			conn.WriteJSON(progress)
			time.Sleep(500 * time.Millisecond) // Simulate processing time
		}
	}

	// Process through coordinator
	err := s.coordinator.ProcessUserRequest(req.Message)

	if err != nil {
		errorResp := ChatResponse{
			Type:      "error",
			Content:   "Failed to process request: " + err.Error(),
			SessionID: session.ID,
			ProjectID: projectID,
			Timestamp: time.Now(),
			Status:    "error",
		}
		conn.WriteJSON(errorResp)
		return
	}

	// Send completion status
	completion := ChatResponse{
		Type:      "completion",
		Content:   "Request processing completed successfully!",
		SessionID: session.ID,
		ProjectID: projectID,
		Timestamp: time.Now(),
		Status:    "completed",
		Progress:  100,
	}
	conn.WriteJSON(completion)

	// Add completion message to session
	completionMsg := ChatMessage{
		ID:        uuid.New().String(),
		Type:      "assistant",
		Content:   "Request processing completed successfully!",
		Timestamp: time.Now(),
		Status:    "completed",
	}
	session.mutex.Lock()
	session.Messages = append(session.Messages, completionMsg)
	session.mutex.Unlock()
}

// handleStatusRequest returns current system status
func (s *Server) handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("[STATUS_REQUEST] timestamp=%s", time.Now().Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"agents":    s.coordinator.ListActiveAgents(),
		"project":   s.coordinator.GetProjectStatus(),
	}

	json.NewEncoder(w).Encode(status)
}

// handleAgentsList returns the list of available agents
func (s *Server) handleAgentsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	agents := s.coordinator.ListActiveAgents()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}

// handleModelsRequest returns available models for the current LLM provider
func (s *Server) handleModelsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	models := s.coordinator.GetAvailableModels()
	provider := s.coordinator.GetLLMProvider()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider": provider,
		"models":   models,
	})
}

// isProjectDirectory checks if a directory contains typical project files
func isProjectDirectory(dir string) bool {
	// Check for common project files
	indicators := []string{"package.json", "go.mod", "requirements.txt", "Cargo.toml", "pom.xml"}

	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
			return true
		}
	}
	return false
}

// handleProjectsRequest returns list of created projects
func (s *Server) handleProjectsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var projects []map[string]interface{}

	// Try to get projects from database first
	if s.projectDB != nil {
		dbProjects, err := s.projectDB.ListProjects()
		if err == nil {
			for _, project := range dbProjects {
				projects = append(projects, map[string]interface{}{
					"name":        project.Name,
					"path":        filepath.Join(s.projectPath, project.Name),
					"status":      project.Status,
					"port":        project.Port,
					"template":    project.Template,
					"container":   project.DockerContainer,
					"created_at":  project.CreatedAt.Format(time.RFC3339),
					"last_update": project.UpdatedAt.Format(time.RFC3339),
				})
			}
		}
	}

	// Fallback: scan actual project directories if database is empty
	if len(projects) == 0 {
		if entries, err := os.ReadDir(s.projectPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					projectDir := filepath.Join(s.projectPath, entry.Name())

					// Check if it looks like a project directory
					if isProjectDirectory(projectDir) {
						stat, _ := entry.Info()
						projects = append(projects, map[string]interface{}{
							"name":        entry.Name(),
							"path":        projectDir,
							"status":      "unknown", // Status unknown for non-DB projects
							"port":        nil,
							"template":    "unknown",
							"container":   nil,
							"created_at":  stat.ModTime().Format(time.RFC3339),
							"last_update": stat.ModTime().Format(time.RFC3339),
						})
					}
				}
			}
		}
	}

	// If no projects found, return empty list
	if len(projects) == 0 {
		projects = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
}

// handleCreateProject handles project creation requests
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var request struct {
		Name        string `json:"name"`
		Template    string `json:"template"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate template
	templates, err := s.templateManager.GetAvailableTemplates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get templates",
		})
		return
	}

	templateExists := false
	for _, template := range templates {
		if template.Name == request.Template {
			templateExists = true
			break
		}
	}

	if !templateExists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid template",
		})
		return
	}

	// Generate project name if not provided or sanitize if provided
	projectName := request.Name
	if projectName == "" {
		projectName = s.templateManager.GenerateProjectName("project")
	} else {
		projectName = s.templateManager.GenerateProjectName(projectName)
	}

	// Copy template
	err = s.templateManager.CopyTemplate(request.Template, projectName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to create project: %v", err),
		})
		return
	}

	// Create database record
	port := 3000 // Default port for React/Next.js
	dockerContainer := projectName

	var project *database.Project
	if s.projectDB != nil {
		project, err = s.projectDB.CreateProject(projectName, request.Template, dockerContainer, port)
		if err != nil {
			log.Printf("Failed to create project in database: %v", err)
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"name":       projectName,
		"template":   request.Template,
		"path":       s.templateManager.GetProjectPath(projectName),
		"container":  dockerContainer,
		"port":       port,
		"status":     "created",
		"created_at": time.Now().Format(time.RFC3339),
	}

	if project != nil {
		response["id"] = project.ID
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("[PROJECT_CREATE] name=%s template=%s container=%s", projectName, request.Template, dockerContainer)
}

// handleTemplatesList returns available templates
func (s *Server) handleTemplatesList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	templates, err := s.templateManager.GetAvailableTemplates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get templates",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// handleProjectPreview handles project preview requests
func (s *Server) handleProjectPreview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["name"]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// For demo purposes, return a preview URL
	// In real implementation, this would start the project and return the actual URL
	previewURL := "http://localhost:3000" // Default React dev server

	json.NewEncoder(w).Encode(map[string]interface{}{
		"project_name": projectName,
		"preview_url":  previewURL,
		"status":       "running",
		"host_path":    fmt.Sprintf("./projects/%s", projectName),
	})
}

// handleHealthCheck returns health status
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	json.NewEncoder(w).Encode(health)
}

// Start begins the HTTP server
func (s *Server) Start() error {
	log.Printf("[SERVER_START] port=%s timestamp=%s", s.port, time.Now().Format(time.RFC3339))

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("[SERVER_LISTENING] address=http://localhost:%s", s.port)
	return server.ListenAndServe()
}

// handleCancelRequest handles chat cancellation requests
func (s *Server) handleCancelRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	w.Header().Set("Content-Type", "application/json")

	s.sessionMutex.Lock()
	session, exists := s.activeSessions[sessionID]
	if exists {
		session.Cancel()
		delete(s.activeSessions, sessionID)
	}
	s.sessionMutex.Unlock()

	response := map[string]interface{}{
		"success":   exists,
		"sessionId": sessionID,
		"message":   "Session cancelled",
	}

	if !exists {
		response["message"] = "Session not found or already completed"
		w.WriteHeader(http.StatusNotFound)
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("[CANCEL_REQUEST] session_id=%s found=%v", sessionID, exists)
}

// handleGetChatSession returns a specific chat session
func (s *Server) handleGetChatSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	w.Header().Set("Content-Type", "application/json")

	s.sessionMutex.RLock()
	session, exists := s.activeSessions[sessionID]
	s.sessionMutex.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Session not found",
		})
		return
	}

	session.mutex.RLock()
	response := map[string]interface{}{
		"id":            session.ID,
		"project_id":    session.ProjectID,
		"created_at":    session.CreatedAt,
		"last_activity": session.LastActivity,
		"messages":      session.Messages,
	}
	session.mutex.RUnlock()

	json.NewEncoder(w).Encode(response)
}

// handleListChatSessions returns all active chat sessions
func (s *Server) handleListChatSessions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	s.sessionMutex.RLock()
	sessions := make([]map[string]interface{}, 0, len(s.activeSessions))
	for _, session := range s.activeSessions {
		session.mutex.RLock()
		sessionInfo := map[string]interface{}{
			"id":            session.ID,
			"project_id":    session.ProjectID,
			"created_at":    session.CreatedAt,
			"last_activity": session.LastActivity,
			"message_count": len(session.Messages),
		}
		session.mutex.RUnlock()
		sessions = append(sessions, sessionInfo)
	}
	s.sessionMutex.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// handleGetProject returns a specific project by ID
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	w.Header().Set("Content-Type", "application/json")

	if s.projectDB == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database not available",
		})
		return
	}

	// For now, try to get project by name (since we're using project names as IDs)
	project, err := s.projectDB.GetProjectByName(projectID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Project not found",
		})
		return
	}

	json.NewEncoder(w).Encode(project)
}

// FileNode represents a file or directory in the project
type FileNode struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"` // "file" or "folder"
	Path     string     `json:"path"`
	Size     int64      `json:"size,omitempty"`
	Children []FileNode `json:"children,omitempty"`
}

// handleProjectFiles returns the file tree structure for a project
func (s *Server) handleProjectFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	projectName := vars["name"]

	// Get project path from server configuration
	projectPath := filepath.Join(s.projectPath, projectName)

	// Check if project directory exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Project not found",
		})
		return
	}

	// Build file tree
	fileTree, err := s.buildFileTree(projectPath, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to read project files: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"project": projectName,
		"files":   fileTree,
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("[PROJECT_FILES] project=%s files_count=%d", projectName, len(fileTree))
}

// handleFileContent handles reading and writing file content
func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	projectName := vars["name"]
	filePath := vars["filepath"]

	// Sanitize file path to prevent directory traversal
	filePath = strings.ReplaceAll(filePath, "..", "")
	projectPath := filepath.Join(s.projectPath, projectName)
	fullPath := filepath.Join(projectPath, filePath)

	// Ensure the file is within the project directory
	if !strings.HasPrefix(fullPath, projectPath) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid file path",
		})
		return
	}

	switch r.Method {
	case "GET":
		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "File not found: " + err.Error(),
			})
			return
		}

		response := map[string]interface{}{
			"file":    filePath,
			"content": string(content),
			"size":    len(content),
		}

		json.NewEncoder(w).Encode(response)
		log.Printf("[FILE_READ] project=%s file=%s size=%d", projectName, filePath, len(content))

	case "POST":
		// Write file content
		var request struct {
			Content string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to create directory: " + err.Error(),
			})
			return
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(request.Content), 0644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to write file: " + err.Error(),
			})
			return
		}

		response := map[string]interface{}{
			"file":    filePath,
			"success": true,
			"size":    len(request.Content),
		}

		json.NewEncoder(w).Encode(response)
		log.Printf("[FILE_WRITE] project=%s file=%s size=%d", projectName, filePath, len(request.Content))
	}
}

// buildFileTree recursively builds a file tree structure
func (s *Server) buildFileTree(basePath, relativePath string) ([]FileNode, error) {
	fullPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var nodes []FileNode
	for _, entry := range entries {
		// Skip hidden files and common ignore patterns
		if strings.HasPrefix(entry.Name(), ".") ||
			entry.Name() == "node_modules" ||
			entry.Name() == "dist" ||
			entry.Name() == "build" {
			continue
		}

		entryPath := filepath.Join(relativePath, entry.Name())
		node := FileNode{
			Name: entry.Name(),
			Path: entryPath,
		}

		if entry.IsDir() {
			node.Type = "folder"
			// Recursively get children (limit depth to prevent infinite recursion)
			if strings.Count(entryPath, string(filepath.Separator)) < 10 {
				children, err := s.buildFileTree(basePath, entryPath)
				if err == nil {
					node.Children = children
				}
			}
		} else {
			node.Type = "file"
			if info, err := entry.Info(); err == nil {
				node.Size = info.Size()
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	return uuid.New().String()
}
