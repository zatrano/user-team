package configs

import (
	"os"
	"strconv"
	"time"

	"zatrano/utils"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		utils.SLog.Warnw(".env dosyası yüklenemedi, sistem ortam değişkenleri kullanılacak (eğer varsa)", "error", err)
	} else {
		utils.SLog.Info(".env dosyası başarıyla yüklendi")
	}

	portStr := utils.GetEnvWithDefault("DB_PORT", "5432")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.SLog.Fatalw("Invalid DB_PORT environment variable",
			"value", portStr,
			"error", err,
		)
	}

	dbConfig := DatabaseConfig{
		Host:     utils.GetEnvWithDefault("DB_HOST", "localhost"),
		Port:     port,
		User:     utils.GetEnvWithDefault("DB_USERNAME", "postgres"),
		Password: utils.GetEnvWithDefault("DB_PASSWORD", ""),
		Name:     utils.GetEnvWithDefault("DB_DATABASE", "myapp"),
		SSLMode:  utils.GetEnvWithDefault("DB_SSL_MODE", "disable"),
		TimeZone: utils.GetEnvWithDefault("DB_TIMEZONE", "UTC"),
	}

	utils.Log.Info("Database configuration loaded",
		zap.String("host", dbConfig.Host),
		zap.Int("port", dbConfig.Port),
		zap.String("user", dbConfig.User),
		zap.String("database", dbConfig.Name),
		zap.String("sslmode", dbConfig.SSLMode),
		zap.String("timezone", dbConfig.TimeZone),
	)

	dsn := "host=" + dbConfig.Host +
		" user=" + dbConfig.User +
		" password=" + dbConfig.Password +
		" dbname=" + dbConfig.Name +
		" port=" + strconv.Itoa(dbConfig.Port) +
		" sslmode=" + dbConfig.SSLMode +
		" TimeZone=" + dbConfig.TimeZone

	var gormerr error
	DB, gormerr = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(getGormLogLevel()),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if gormerr != nil {
		utils.Log.Fatal("Failed to connect to database",
			zap.String("host", dbConfig.Host),
			zap.Int("port", dbConfig.Port),
			zap.String("user", dbConfig.User),
			zap.String("database", dbConfig.Name),
			zap.Error(gormerr),
		)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		utils.Log.Fatal("Failed to get underlying sql.DB instance", zap.Error(err))
	}

	maxIdleConns := utils.GetEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	maxOpenConns := utils.GetEnvAsInt("DB_MAX_OPEN_CONNS", 100)
	connMaxLifetimeMinutes := utils.GetEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 60)

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetimeMinutes) * time.Minute)

	utils.Log.Info("Database connection established successfully",
		zap.Int("max_idle_conns", maxIdleConns),
		zap.Int("max_open_conns", maxOpenConns),
		zap.Int("conn_max_lifetime_minutes", connMaxLifetimeMinutes),
	)
}

func getGormLogLevel() logger.LogLevel {
	switch os.Getenv("DB_LOG_LEVEL") {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}

func GetDB() *gorm.DB {
	if DB == nil {
		utils.Log.Fatal("Database connection not initialized. Call InitDB() first.")
	}
	return DB
}

func CloseDB() error {
	if DB == nil {
		utils.SLog.Info("Database connection already closed or not initialized.")
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		utils.Log.Error("Failed to get database instance for closing", zap.Error(err))
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		utils.Log.Error("Error closing database connection", zap.Error(err))
		return err
	}

	utils.SLog.Info("Database connection closed successfully.")
	DB = nil
	return nil
}
