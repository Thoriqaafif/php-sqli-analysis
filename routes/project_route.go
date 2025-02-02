package routes

import (
	"github.com/Thoriqaafif/php-sqli-analysis/controller"
	"github.com/labstack/echo/v4"
)

func Project(route *echo.Echo, projectController controller.ProjectController) {
	routes := route.Group("/api/project")
	{
		routes.GET("", projectController.GetProjects)
		routes.POST("", projectController.CreateProject)
		routes.GET("/:id", projectController.GetProject)
		routes.DELETE("/delete/:id", projectController.DeleteProject)
	}
}
