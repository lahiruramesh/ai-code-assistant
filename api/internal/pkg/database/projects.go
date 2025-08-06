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
