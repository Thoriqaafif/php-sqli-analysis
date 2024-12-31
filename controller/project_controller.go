package controller

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type ProjectController interface {
	GetProjects(ctx echo.Context) error
	CreateProject(ctx echo.Context) error
	GetProject(ctx echo.Context) error
	GetScanResult(ctx echo.Context) error
	GetScannedFile(ctx echo.Context) error
	AddScanResult(ctx echo.Context) error
}

type projectController struct {
	db *pgx.Conn
}

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	NumOfFiles    int       `json:"num_of_files"`
	DetectedVulns int       `json:"detected_vulns"`
	CreatedAt     time.Time `json:"created_at"`
	Scanned       bool      `json:"scanned"`
}

func NewProjectController(db *pgx.Conn) ProjectController {
	return &projectController{
		db: db,
	}
}

func (c *projectController) GetProjects(ctx echo.Context) error {

}

func (c *projectController) CreateProject(ctx echo.Context) error {

}

func (c *projectController) GetProject(ctx echo.Context) error {

}

// return sqli scanned result based on id
func (c *projectController) GetScanResult(ctx echo.Context) error {

}

func (c *projectController) GetScannedFile(ctx echo.Context) error {

}

// receive a sqli scan result and store to the storage
func (c *projectController) AddScanResult(ctx echo.Context) error {

}
