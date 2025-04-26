package middlewares

import (
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

func GuestMiddleware(c *fiber.Ctx) error {
	sess, err := utils.SessionStart(c)
	if err != nil {
		return c.Next()
	}

	userID, err := utils.GetUserIDFromSession(sess)
	if err != nil {
		return c.Next()
	}

	authService := services.NewAuthService()
	user, err := authService.GetUserProfile(userID)
	if err != nil {
		_ = sess.Destroy()
		return c.Next()
	}

	var redirectURL string
	switch user.Type {
	case models.Manager:
		redirectURL = "/manager/home"
	case models.Agent:
		redirectURL = "/agent/home"
	case models.System:
		redirectURL = "/dashboard/home"
	default:
		_ = sess.Destroy()
		return c.Next()
	}

	return c.Redirect(redirectURL)
}
