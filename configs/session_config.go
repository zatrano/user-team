package configs

import (
	"encoding/gob"
	"strconv"
	"time"

	"zatrano/models"
	"zatrano/utils"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
	"go.uber.org/zap"
)

type sessionDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

var Session *session.Store

func InitSession() {
	Session = createSessionStore()
	utils.InitializeSessionStore(Session)
	utils.SLog.Info("Session store initialized and registered in utils")
}

func SetupSession() *session.Store {
	if Session == nil {
		utils.SLog.Warn("Session store requested but not initialized, initializing now.")
		InitSession()
	}
	return Session
}

func createSessionStore() *session.Store {
	portStr := utils.GetEnvWithDefault("DB_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.SLog.Fatalw("Invalid DB_PORT environment variable for session store",
			"value", portStr,
			"error", err,
		)
	}

	dbConfig := sessionDatabaseConfig{
		Host:     utils.GetEnvWithDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     utils.GetEnvWithDefault("DB_USERNAME", "postgres"),
		Password: utils.GetEnvWithDefault("DB_PASSWORD", ""),
		Name:     utils.GetEnvWithDefault("DB_DATABASE", "myapp"),
		SSLMode:  utils.GetEnvWithDefault("DB_SSL_MODE", "disable"),
	}

	utils.Log.Info("Configuring session storage",
		zap.String("storage_type", "postgres"),
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("user", dbConfig.User),
		zap.String("database", dbConfig.Name),
		zap.String("sslmode", dbConfig.SSLMode),
		zap.String("table", "sessions"),
	)

	storage := postgres.New(postgres.Config{
		Host:       dbConfig.Host,
		Port:       dbConfig.Port,
		Username:   dbConfig.User,
		Password:   dbConfig.Password,
		Database:   dbConfig.Name,
		SSLMode:    dbConfig.SSLMode,
		Reset:      false,
		Table:      "sessions",
		GCInterval: 10 * time.Second,
	})

	sessionExpirationHours := utils.GetEnvAsInt("SESSION_EXPIRATION_HOURS", 24)
	store := session.New(session.Config{
		Storage:    storage,
		Expiration: time.Duration(sessionExpirationHours) * time.Hour,
	})

	utils.SLog.Infof("Session store configured with %d hour expiration", sessionExpirationHours)

	registerGobTypes()

	return store
}

func registerGobTypes() {
	gob.Register(models.UserType(""))
	gob.Register(&models.User{})
	utils.SLog.Debug("Registered gob types for session: models.UserType, *models.User")
}
