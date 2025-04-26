package routes

import (
	"zatrano/configs"
	"zatrano/models"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	app.Use(logger.New())

	sessionStore := configs.SetupSession()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("session", sessionStore)
		return c.Next()
	})

	registerAuthRoutes(app)
	registerDashboardRoutes(app)
	registerManagerRoutes(app)
	registerAgentRoutes(app)

	app.Use(rootRedirector)
}

func rootRedirector(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)
	if err != nil {
		return c.Redirect("/auth/login")
	}

	_, err = utils.GetUserIDFromSession(sess)
	if err != nil {
		return c.Redirect("/auth/login")
	}

	userType, err := utils.GetUserTypeFromSession(sess)
	if err != nil {
		return c.Redirect("/auth/login")
	}

	switch userType {
	case models.Manager:
		return c.Redirect("/manager/home")
	case models.Agent:
		return c.Redirect("/agent/home")
	case models.System:
		return c.Redirect("/dashboard/home")
	default:
		return c.SendString("Geçersiz kullanıcı tipi")
	}
}
