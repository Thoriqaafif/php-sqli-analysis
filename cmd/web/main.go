package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Thoriqaafif/php-sqli-analysis/controller"
	"github.com/Thoriqaafif/php-sqli-analysis/routes"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	addr := host + ":" + port

	// config db
	db, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close(context.Background())

	server := echo.New()

	// middleware
	// Enable CORS middleware
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},                                                // Allowed origins
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete}, // Allowed HTTP methods
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// controllers
	projectController := controller.NewProjectController(db)

	// routes
	routes.Project(server, projectController)

	// register static file
	server.Static("/", wd+"/web/dist")
	server.File("/", wd+"/web/dist/index.html")

	server.Logger.Fatal(server.Start(addr))
}
