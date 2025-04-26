package middlewares

import (
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)

	if err != nil {
		return c.Redirect("/auth/login")
	}

	userID, err := utils.GetUserIDFromSession(sess)

	if err != nil {
		return c.Redirect("/auth/login")
	}

	authService := services.NewAuthService()

	_, err = authService.GetUserProfile(userID)

	if err != nil {
		_ = sess.Destroy()
		return c.Redirect("/auth/login")
	}

	return c.Next()
}
