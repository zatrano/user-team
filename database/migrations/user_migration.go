package migrations

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateUsersTable(db *gorm.DB) error {
	err := db.Exec(`DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
			CREATE TYPE user_type AS ENUM ('system', 'manager', 'agent');
		END IF;
	END$$`).Error
	if err != nil {
		utils.Log.Error("Failed to create/check user_type enum", zap.Error(err))
		return err
	}
	utils.SLog.Debug("Checked/created user_type enum")

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		utils.Log.Error("Failed to migrate users table structure", zap.Error(err))
		return err
	}
	utils.SLog.Info("Users table structure migrated successfully")

	constraintName := "fk_users_team"
	if !db.Migrator().HasConstraint(&models.User{}, constraintName) {
		utils.SLog.Debugf("Constraint %s not found, attempting to add", constraintName)
		err = db.Exec(`
			ALTER TABLE users
			ADD CONSTRAINT fk_users_team
			FOREIGN KEY (team_id) REFERENCES teams(id)
			ON UPDATE CASCADE ON DELETE SET NULL
		`).Error

		if err != nil {
			if !db.Migrator().HasConstraint(&models.User{}, constraintName) {
				utils.Log.Error("Failed to add team foreign key constraint", zap.String("constraint", constraintName), zap.Error(err))
				return err
			} else {
				utils.SLog.Warnf("Adding constraint %s reported an error, but constraint now exists", constraintName, zap.Error(err))
			}
		} else {
			utils.SLog.Infof("Added constraint %s successfully", constraintName)
		}
	} else {
		utils.SLog.Infof("Constraint %s already exists", constraintName)
	}

	return nil
}
