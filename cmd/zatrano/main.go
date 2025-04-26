package main

import (
	"os"
	"os/signal"
	"syscall"

	"zatrano/configs"
	"zatrano/routes"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	utils.InitLogger()
	defer utils.SyncLogger()

	utils.SLog.Debugw("Ortam değişkenleri yüklendi ve logger başlatıldı")

	configs.InitDB()
	defer configs.CloseDB()

	configs.InitSession()

	engine := html.New("./views", ".html")
	engine.AddFunc("getFlashMessages", utils.GetFlashMessages)
	engine.AddFuncMap(utils.TemplateHelpers())

	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			utils.Log.Error("Fiber request error",
				zap.Error(err),
				zap.Int("status_code", code),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
			)

			return c.Status(code).JSON(fiber.Map{"error": message})
		},
	})

	app.Static("/", "./public")
	app.Use(configs.SetupCSRF())
	routes.SetupRoutes(app, configs.GetDB())

	startServer(app)
}

func startServer(app *fiber.App) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		port := os.Getenv("APP_PORT")
		if port == "" {
			port = "3000"
		}
		address := ":" + port
		fullAddress := "http://localhost" + address

		utils.Log.Info("Uygulama başlatılıyor",
			zap.String("address", fullAddress),
			zap.String("port", port),
		)

		if err := app.Listen(address); err != nil {
			utils.Log.Fatal("Sunucu dinlenemedi",
				zap.String("address", address),
				zap.Error(err),
			)
		}
	}()

	<-shutdown
	utils.Log.Info("Kapatma sinyali alındı, uygulama kapatılıyor...")

	if err := app.Shutdown(); err != nil {
		utils.Log.Error("Sunucu kapatılırken hata oluştu", zap.Error(err))
	} else {
		utils.Log.Info("Sunucu başarıyla kapatıldı")
	}

	utils.Log.Info("Uygulama başarıyla sonlandırıldı.")
}
