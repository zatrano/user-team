package configs

import (
	"time"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	fiberUtils "github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

func SetupCSRF() fiber.Handler {
	config := csrf.Config{
		KeyLookup:      "form:csrf_token",
		CookieName:     "csrf_",
		CookieHTTPOnly: true,
		CookieSecure:   false,
		CookieSameSite: "Lax",
		Expiration:     1 * time.Hour,
		KeyGenerator:   fiberUtils.UUID,
		ContextKey:     "csrf",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			utils.Log.Warn("CSRF validation failed",
				zap.Error(err),
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
			_ = utils.SetFlashMessage(c, utils.FlashErrorKey, "Geçersiz işlem. Lütfen sayfayı yenileyin.")
			return c.RedirectBack("/auth/login")
		},
	}

	utils.SLog.Info("CSRF middleware yapılandırıldı")
	return csrf.New(config)
}
