package middlewares

import (
	"zatrano/models"
	"zatrano/services"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
)

func TypeMiddleware(requiredType models.UserType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := utils.SessionStart(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Oturum açılmamış")
		}

		userID, err := utils.GetUserIDFromSession(sess)
		if err != nil {
			return c.Status(fiber.StatusForbidden).SendString("Yetkisiz erişim")
		}

		authService := services.NewAuthService()
		user, err := authService.GetUserProfile(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Kullanıcı bilgileri alınamadı")
		}

		if user.Type != requiredType {
			return c.Status(fiber.StatusForbidden).SendString("Bu işlem için yetkiniz yok")
		}

		return c.Next()
	}
}
