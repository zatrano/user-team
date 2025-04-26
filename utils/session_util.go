package utils

import (
	"zatrano/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var store *session.Store

func InitializeSessionStore(s *session.Store) {
	store = s
}

func SessionStart(c *fiber.Ctx) (*session.Session, error) {
	if store == nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "session store not initialized")
	}
	return store.Get(c)
}

func GetUserTypeFromSession(sess *session.Session) (models.UserType, error) {
	if ut, ok := sess.Get("user_type").(models.UserType); ok {
		return ut, nil
	}

	if utStr, ok := sess.Get("user_type").(string); ok {
		return models.UserType(utStr), nil
	}

	return "", fiber.NewError(fiber.StatusUnauthorized, "Geçersiz oturum veya kullanıcı tipi")
}

func GetUserIDFromSession(sess *session.Session) (uint, error) {
	userID, ok := sess.Get("user_id").(uint)
	if !ok {
		return 0, fiber.NewError(fiber.StatusUnauthorized, "Geçersiz oturum veya kullanıcı ID'si")
	}
	return userID, nil
}
