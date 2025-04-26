package routes

import (
	handlers "zatrano/handlers/agent"
	"zatrano/middlewares"
	"zatrano/models"

	"github.com/gofiber/fiber/v2"
)

func registerAgentRoutes(app *fiber.App) {
	agentGroup := app.Group("/agent")
	agentGroup.Use(
		middlewares.AuthMiddleware,
		middlewares.StatusMiddleware,
		middlewares.TypeMiddleware(models.Agent),
	)

	agentGroup.Get("/home", handlers.AgentHomeHandler)
}
