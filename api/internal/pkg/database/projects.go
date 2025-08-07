package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ProjectDB handles all database operations for project tracking
type ProjectDB struct {
	db *sql.DB
}

// Project represents a project in the database
type Project struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Template        string    `json:"template"`
	DockerContainer string    `json:"docker_container"`
	Port            int       `json:"port"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Container represents a Docker container in the database
type Container struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ProjectID   int       `json:"project_id"`
	Status      string    `json:"status"`
	PortMapping string    `json:"port_mapping"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenUsage represents token usage tracking
type TokenUsage struct {
	ID           int       `json:"id"`
	SessionID    string    `json:"session_id"`
	ProjectID    *int      `json:"project_id,omitempty"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	RequestType  string    `json:"request_type"` // chat, generation, etc.
	CreatedAt    time.Time `json:"created_at"`
}

// ConversationMessage represents a message in a conversation
type ConversationMessage struct {
	ID           int       `json:"id"`
	SessionID    string    `json:"session_id"`
	ProjectID    *int      `json:"project_id,omitempty"`
	Role         string    `json:"role"` // user, assistant, system
	Content      string    `json:"content"`
	Model        string    `json:"model,omitempty"`
	Provider     string    `json:"provider,omitempty"`
	TokenUsageID *int      `json:"token_usage_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewProjectDB creates a new database connection and initializes tables
func NewProjectDB(dbPath string) (*ProjectDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	pdb := &ProjectDB{db: db}

	// Initialize tables
	if err := pdb.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return pdb, nil
}

// initTables creates the necessary tables if they don't exist
func (pdb *ProjectDB) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			template TEXT NOT NULL,
			docker_container TEXT UNIQUE,
			port INTEGER,
			status TEXT DEFAULT 'created',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			project_id INTEGER,
			status TEXT DEFAULT 'created',
			port_mapping TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects (id)
		)`,
		`CREATE TABLE IF NOT EXISTS token_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			project_id INTEGER,
			model TEXT NOT NULL,
			provider TEXT NOT NULL,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			total_tokens INTEGER DEFAULT 0,
			request_type TEXT DEFAULT 'chat',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects (id)
		)`,
		`CREATE TABLE IF NOT EXISTS conversation_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			project_id INTEGER,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			model TEXT,
			provider TEXT,
			token_usage_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects (id),
			FOREIGN KEY (token_usage_id) REFERENCES token_usage (id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_token_usage_session ON token_usage(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_token_usage_project ON token_usage(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_session ON conversation_messages(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_project ON conversation_messages(project_id)`,
	}

	for _, query := range queries {
		if _, err := pdb.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %v", query, err)
		}
	}

	return nil
}

// CreateProject creates a new project in the database
func (pdb *ProjectDB) CreateProject(name, template, dockerContainer string, port int) (*Project, error) {
	query := `INSERT INTO projects (name, template, docker_container, port, status) 
			  VALUES (?, ?, ?, ?, 'created')`

	result, err := pdb.db.Exec(query, name, template, dockerContainer, port)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID: %v", err)
	}

	return pdb.GetProject(int(id))
}

// GetProject retrieves a project by ID
func (pdb *ProjectDB) GetProject(id int) (*Project, error) {
	query := `SELECT id, name, template, docker_container, port, status, created_at, updated_at 
			  FROM projects WHERE id = ?`

	var p Project
	err := pdb.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Template, &p.DockerContainer,
		&p.Port, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	return &p, nil
}

// GetProjectByName retrieves a project by name
func (pdb *ProjectDB) GetProjectByName(name string) (*Project, error) {
	query := `SELECT id, name, template, docker_container, port, status, created_at, updated_at 
			  FROM projects WHERE name = ?`

	var p Project
	err := pdb.db.QueryRow(query, name).Scan(
		&p.ID, &p.Name, &p.Template, &p.DockerContainer,
		&p.Port, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	return &p, nil
}

// UpdateProjectStatus updates the status of a project
func (pdb *ProjectDB) UpdateProjectStatus(id int, status string) error {
	query := `UPDATE projects SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err := pdb.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update project status: %v", err)
	}

	return nil
}

// ListProjects returns all projects
func (pdb *ProjectDB) ListProjects() ([]Project, error) {
	query := `SELECT id, name, template, docker_container, port, status, created_at, updated_at 
			  FROM projects ORDER BY created_at DESC`

	rows, err := pdb.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID, &p.Name, &p.Template, &p.DockerContainer,
			&p.Port, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// CreateContainer creates a new container record
func (pdb *ProjectDB) CreateContainer(name string, projectID int, portMapping string) (*Container, error) {
	query := `INSERT INTO containers (name, project_id, port_mapping, status) 
			  VALUES (?, ?, ?, 'created')`

	result, err := pdb.db.Exec(query, name, projectID, portMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get container ID: %v", err)
	}

	return pdb.GetContainer(int(id))
}

// GetContainer retrieves a container by ID
func (pdb *ProjectDB) GetContainer(id int) (*Container, error) {
	query := `SELECT id, name, project_id, status, port_mapping, created_at 
			  FROM containers WHERE id = ?`

	var c Container
	err := pdb.db.QueryRow(query, id).Scan(
		&c.ID, &c.Name, &c.ProjectID, &c.Status, &c.PortMapping, &c.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get container: %v", err)
	}

	return &c, nil
}

// UpdateContainerStatus updates the status of a container
func (pdb *ProjectDB) UpdateContainerStatus(id int, status string) error {
	query := `UPDATE containers SET status = ? WHERE id = ?`

	_, err := pdb.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update container status: %v", err)
	}

	return nil
}

// Close closes the database connection
func (pdb *ProjectDB) Close() error {
	return pdb.db.Close()
}

// CreateTokenUsage creates a new token usage record
func (pdb *ProjectDB) CreateTokenUsage(sessionID string, projectID *int, model, provider string, inputTokens, outputTokens int, requestType string) (*TokenUsage, error) {
	totalTokens := inputTokens + outputTokens
	query := `INSERT INTO token_usage (session_id, project_id, model, provider, input_tokens, output_tokens, total_tokens, request_type) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := pdb.db.Exec(query, sessionID, projectID, model, provider, inputTokens, outputTokens, totalTokens, requestType)
	if err != nil {
		return nil, fmt.Errorf("failed to create token usage: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage ID: %v", err)
	}

	return pdb.GetTokenUsage(int(id))
}

// GetTokenUsage retrieves a token usage record by ID
func (pdb *ProjectDB) GetTokenUsage(id int) (*TokenUsage, error) {
	query := `SELECT id, session_id, project_id, model, provider, input_tokens, output_tokens, total_tokens, request_type, created_at 
			  FROM token_usage WHERE id = ?`

	var tu TokenUsage
	var projectID sql.NullInt64
	err := pdb.db.QueryRow(query, id).Scan(
		&tu.ID, &tu.SessionID, &projectID, &tu.Model, &tu.Provider,
		&tu.InputTokens, &tu.OutputTokens, &tu.TotalTokens, &tu.RequestType, &tu.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get token usage: %v", err)
	}

	if projectID.Valid {
		pid := int(projectID.Int64)
		tu.ProjectID = &pid
	}

	return &tu, nil
}

// GetSessionTokenUsage retrieves all token usage for a session
func (pdb *ProjectDB) GetSessionTokenUsage(sessionID string) ([]TokenUsage, error) {
	query := `SELECT id, session_id, project_id, model, provider, input_tokens, output_tokens, total_tokens, request_type, created_at 
			  FROM token_usage WHERE session_id = ? ORDER BY created_at ASC`

	rows, err := pdb.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session token usage: %v", err)
	}
	defer rows.Close()

	var usages []TokenUsage
	for rows.Next() {
		var tu TokenUsage
		var projectID sql.NullInt64
		err := rows.Scan(
			&tu.ID, &tu.SessionID, &projectID, &tu.Model, &tu.Provider,
			&tu.InputTokens, &tu.OutputTokens, &tu.TotalTokens, &tu.RequestType, &tu.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %v", err)
		}

		if projectID.Valid {
			pid := int(projectID.Int64)
			tu.ProjectID = &pid
		}

		usages = append(usages, tu)
	}

	return usages, nil
}

// GetProjectTokenUsage retrieves all token usage for a project
func (pdb *ProjectDB) GetProjectTokenUsage(projectID int) ([]TokenUsage, error) {
	query := `SELECT id, session_id, project_id, model, provider, input_tokens, output_tokens, total_tokens, request_type, created_at 
			  FROM token_usage WHERE project_id = ? ORDER BY created_at DESC`

	rows, err := pdb.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project token usage: %v", err)
	}
	defer rows.Close()

	var usages []TokenUsage
	for rows.Next() {
		var tu TokenUsage
		var projectID sql.NullInt64
		err := rows.Scan(
			&tu.ID, &tu.SessionID, &projectID, &tu.Model, &tu.Provider,
			&tu.InputTokens, &tu.OutputTokens, &tu.TotalTokens, &tu.RequestType, &tu.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %v", err)
		}

		if projectID.Valid {
			pid := int(projectID.Int64)
			tu.ProjectID = &pid
		}

		usages = append(usages, tu)
	}

	return usages, nil
}

// CreateConversationMessage creates a new conversation message
func (pdb *ProjectDB) CreateConversationMessage(sessionID string, projectID *int, role, content, model, provider string, tokenUsageID *int) (*ConversationMessage, error) {
	query := `INSERT INTO conversation_messages (session_id, project_id, role, content, model, provider, token_usage_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := pdb.db.Exec(query, sessionID, projectID, role, content, model, provider, tokenUsageID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation message: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation message ID: %v", err)
	}

	return pdb.GetConversationMessage(int(id))
}

// GetConversationMessage retrieves a conversation message by ID
func (pdb *ProjectDB) GetConversationMessage(id int) (*ConversationMessage, error) {
	query := `SELECT id, session_id, project_id, role, content, model, provider, token_usage_id, created_at 
			  FROM conversation_messages WHERE id = ?`

	var cm ConversationMessage
	var projectID, tokenUsageID sql.NullInt64
	var model, provider sql.NullString
	err := pdb.db.QueryRow(query, id).Scan(
		&cm.ID, &cm.SessionID, &projectID, &cm.Role, &cm.Content,
		&model, &provider, &tokenUsageID, &cm.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get conversation message: %v", err)
	}

	if projectID.Valid {
		pid := int(projectID.Int64)
		cm.ProjectID = &pid
	}
	if model.Valid {
		cm.Model = model.String
	}
	if provider.Valid {
		cm.Provider = provider.String
	}
	if tokenUsageID.Valid {
		tuid := int(tokenUsageID.Int64)
		cm.TokenUsageID = &tuid
	}

	return &cm, nil
}

// GetSessionConversation retrieves all messages for a session
func (pdb *ProjectDB) GetSessionConversation(sessionID string) ([]ConversationMessage, error) {
	query := `SELECT id, session_id, project_id, role, content, model, provider, token_usage_id, created_at 
			  FROM conversation_messages WHERE session_id = ? ORDER BY created_at ASC`

	rows, err := pdb.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session conversation: %v", err)
	}
	defer rows.Close()

	var messages []ConversationMessage
	for rows.Next() {
		var cm ConversationMessage
		var projectID, tokenUsageID sql.NullInt64
		var model, provider sql.NullString
		err := rows.Scan(
			&cm.ID, &cm.SessionID, &projectID, &cm.Role, &cm.Content,
			&model, &provider, &tokenUsageID, &cm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation message: %v", err)
		}

		if projectID.Valid {
			pid := int(projectID.Int64)
			cm.ProjectID = &pid
		}
		if model.Valid {
			cm.Model = model.String
		}
		if provider.Valid {
			cm.Provider = provider.String
		}
		if tokenUsageID.Valid {
			tuid := int(tokenUsageID.Int64)
			cm.TokenUsageID = &tuid
		}

		messages = append(messages, cm)
	}

	return messages, nil
}

// GetTokenUsageStats returns aggregated token usage statistics
func (pdb *ProjectDB) GetTokenUsageStats() (map[string]interface{}, error) {
	query := `SELECT 
		COUNT(*) as total_requests,
		SUM(input_tokens) as total_input_tokens,
		SUM(output_tokens) as total_output_tokens,
		SUM(total_tokens) as total_tokens,
		provider,
		model
	FROM token_usage 
	GROUP BY provider, model
	ORDER BY total_tokens DESC`

	rows, err := pdb.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage stats: %v", err)
	}
	defer rows.Close()

	var stats []map[string]interface{}
	var totalRequests, totalInputTokens, totalOutputTokens, totalTokens int

	for rows.Next() {
		var requests, inputTokens, outputTokens, tokens int
		var provider, model string
		err := rows.Scan(&requests, &inputTokens, &outputTokens, &tokens, &provider, &model)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage stats: %v", err)
		}

		stats = append(stats, map[string]interface{}{
			"provider":      provider,
			"model":         model,
			"requests":      requests,
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"total_tokens":  tokens,
		})

		totalRequests += requests
		totalInputTokens += inputTokens
		totalOutputTokens += outputTokens
		totalTokens += tokens
	}

	return map[string]interface{}{
		"total_requests":      totalRequests,
		"total_input_tokens":  totalInputTokens,
		"total_output_tokens": totalOutputTokens,
		"total_tokens":        totalTokens,
		"by_provider_model":   stats,
	}, nil
}
