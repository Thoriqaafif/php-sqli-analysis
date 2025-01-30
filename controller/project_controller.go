package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/report"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type ProjectController interface {
	GetProjects(ctx echo.Context) error
	CreateProject(ctx echo.Context) error
	GetProject(ctx echo.Context) error
	DeleteProject(ctx echo.Context) error
}

type projectController struct {
	db *pgx.Conn
}

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	NumOfFiles    int       `json:"num_of_files"`
	DetectedVulns int       `json:"detected_vulns"`
	ScanTime      float64   `json:"scan_time"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewProjectController(db *pgx.Conn) ProjectController {
	return &projectController{
		db: db,
	}
}

// return projects from database
func (pc *projectController) GetProjects(c echo.Context) error {
	ctx := context.Background()
	q := "SELECT id, name, num_of_files, detected_vulns, scan_time, created_at FROM projects ORDER BY created_at DESC"
	rows, err := pc.db.Query(ctx, q)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Server Error",
		})
	}

	projects, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Project, error) {
		var project Project
		err := row.Scan(&project.ID, &project.Name, &project.NumOfFiles, &project.DetectedVulns,
			&project.ScanTime, &project.CreatedAt)
		return project, err
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Server Error",
		})
	}

	log.Printf("GET /api/project")
	return c.JSON(http.StatusOK, projects)
}

// create new project and receive scan result
func (pc *projectController) CreateProject(c echo.Context) error {
	ctx := context.Background()

	// extract project data
	var newProject struct {
		Data   Project           `json:"data"`
		Result report.ScanReport `json:"result"`
	}
	if err := c.Bind(&newProject); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
	}

	// insert data into db
	q := "INSERT INTO projects (name, num_of_files, detected_vulns, scan_time, created_at) VALUES" +
		" ($1, $2, $3, $4, $5) RETURNING id::text"
	err := pc.db.QueryRow(ctx, q, newProject.Data.Name, newProject.Data.NumOfFiles,
		newProject.Data.DetectedVulns, newProject.Data.ScanTime, newProject.Data.CreatedAt).Scan(&newProject.Data.ID)
	if err != nil {
		log.Printf("POST /api/project: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	// store scan result into storage
	scanResult, err := json.Marshal(newProject.Result)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request scan result"})
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("error: fail to get working directory")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	dirPath := wd + "/storage/scanned_projects/" + newProject.Data.ID
	filePath := dirPath + "/result.json"
	err = os.MkdirAll(dirPath, 0666)
	if err != nil {
		log.Printf("POST /api/project: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	err = os.WriteFile(filePath, scanResult, 0666)
	if err != nil {
		log.Printf("POST /api/project: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"success": "scanned data received"})
}

func (pc *projectController) GetProject(c echo.Context) error {
	ctx := context.Background()
	id := c.Param("id")
	var project struct {
		Data   Project           `json:"data"`
		Result report.ScanReport `json:"result"`
	}

	// get project data
	q := "SELECT id, name, num_of_files, detected_vulns, scan_time, created_at FROM projects WHERE id=$1"
	err := pc.db.QueryRow(ctx, q, id).Scan(&project.Data.ID, &project.Data.Name, &project.Data.NumOfFiles,
		&project.Data.DetectedVulns, &project.Data.ScanTime, &project.Data.CreatedAt)
	if err != nil {
		log.Printf("GET /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	// get project scan result
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("GET /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	filePath := wd + "/storage/scanned_projects/" + id + "/result.json"
	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("GET /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	if err := json.Unmarshal(b, &project.Result); err != nil {
		log.Printf("GET /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	return c.JSON(http.StatusOK, project)
}

// delete project data in db and project scan result in storage
func (pc *projectController) DeleteProject(c echo.Context) error {
	id := c.Param("id")

	// delete project data from db
	q := "DELETE FROM projects WHERE id=$1"
	ret, err := pc.db.Exec(c.Request().Context(), q, id)
	if err != nil {
		log.Printf("DELETE /api/project/delete/:id %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
	if ret.RowsAffected() == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "project not found"})
	}

	// delete project scan result
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("DELETE /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}
	dirPath := wd + "/storage/scanned_projects/" + id
	err = os.RemoveAll(dirPath)
	if err != nil {
		log.Printf("DELETE /api/project/:id: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{})
}
