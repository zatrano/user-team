package routes

import (
	handlers "zatrano/handlers/manager"
	"zatrano/middlewares"
	"zatrano/models"

	"github.com/gofiber/fiber/v2"
)

func registerManagerRoutes(app *fiber.App) {
	managerGroup := app.Group("/manager")
	managerGroup.Use(
		middlewares.AuthMiddleware,
		middlewares.StatusMiddleware,
		middlewares.TypeMiddleware(models.Manager),
	)

	managerGroup.Get("/home", handlers.ManagerHomeHandler)
}
