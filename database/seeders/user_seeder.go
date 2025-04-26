package seeders

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetSystemUserConfig() models.User {
	return models.User{
		Name:     "System",
		Account:  "system@system",
		Type:     models.System,
		Password: "S1st3m@S1st3m",
	}
}

func SeedSystemUser(db *gorm.DB) error {
	systemUserConfig := GetSystemUserConfig()

	userToSeed := models.User{
		Name:     systemUserConfig.Name,
		Account:  systemUserConfig.Account,
		Type:     systemUserConfig.Type,
		Password: systemUserConfig.Password,
		Status:   true,
	}

	var existingUser models.User
	result := db.Where("account = ? AND type = ?", userToSeed.Account, userToSeed.Type).First(&existingUser)

	if result.Error == nil {
		utils.SLog.Infof("Sistem kullanıcısı '%s' zaten mevcut. Güncelleme gerekip gerekmediği kontrol ediliyor...", userToSeed.Account)

		updateFields := make(map[string]interface{})
		needsUpdate := false

		if existingUser.Name != userToSeed.Name {
			updateFields["name"] = userToSeed.Name
			needsUpdate = true
		}
		if !existingUser.Status {
			updateFields["status"] = true
			needsUpdate = true
		}

		if needsUpdate {
			utils.SLog.Infof("Mevcut sistem kullanıcısı '%s' güncelleniyor...", userToSeed.Account)
			err := db.Model(&existingUser).Updates(updateFields).Error
			if err != nil {
				utils.Log.Error("Mevcut sistem kullanıcısı güncellenemedi",
					zap.String("account", userToSeed.Account),
					zap.Error(err),
				)
				return err
			}
			utils.SLog.Infof("Mevcut sistem kullanıcısı '%s' başarıyla güncellendi.", userToSeed.Account)
		} else {
			utils.SLog.Infof("Mevcut sistem kullanıcısı '%s' için güncelleme gerekmiyor.", userToSeed.Account)
		}
		return nil

	} else if result.Error != gorm.ErrRecordNotFound {
		utils.Log.Error("Sistem kullanıcısı kontrol edilirken veritabanı hatası",
			zap.String("account", userToSeed.Account),
			zap.Error(result.Error),
		)
		return result.Error
	}

	utils.SLog.Infof("Sistem kullanıcısı '%s' bulunamadı. Oluşturuluyor...", userToSeed.Account)
	err := db.Create(&userToSeed).Error
	if err != nil {
		utils.Log.Error("Sistem kullanıcısı oluşturulamadı",
			zap.String("account", userToSeed.Account),
			zap.Error(err),
		)
		return err
	}

	utils.SLog.Infof("Sistem kullanıcısı '%s' başarıyla oluşturuldu.", userToSeed.Account)
	return nil
}
