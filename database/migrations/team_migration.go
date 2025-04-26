package migrations

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateTeamsTable(db *gorm.DB) error {
	err := db.AutoMigrate(&models.Team{})
	if err != nil {
		utils.Log.Error("Failed to migrate teams table", zap.Error(err))
		return err
	}

	utils.SLog.Info("Teams table migrated successfully")
	return nil
}
