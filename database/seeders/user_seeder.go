package seeders

import (
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(systemUserConfig.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Log.Error("Sistem kullanıcısı için parola hashlenemedi",
			zap.String("account", systemUserConfig.Account),
			zap.Error(err),
		)
		return err
	}
	hashedPassword := string(hashedPasswordBytes)
	utils.SLog.Debugf("Sistem kullanıcısı '%s' parolası başarıyla hash'lendi", systemUserConfig.Account)

	userToSeed := models.User{
		Name:     systemUserConfig.Name,
		Account:  systemUserConfig.Account,
		Type:     systemUserConfig.Type,
		Password: hashedPassword,
		Status:   true,
	}

	var existingUser models.User
	result := db.Where("account = ? AND type = ?", userToSeed.Account, userToSeed.Type).First(&existingUser)

	if result.Error == nil {
		utils.SLog.Infof("Sistem kullanıcısı '%s' zaten mevcut. Güncelleme gerekip gerekmediği kontrol ediliyor...", userToSeed.Account)

		pwMatchErr := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(systemUserConfig.Password))
		needsUpdate := false
		updateFields := make(map[string]interface{})

		if pwMatchErr != nil {
			utils.SLog.Infof("Sistem kullanıcısı '%s' parolası güncellenmeli.", userToSeed.Account)
			updateFields["password"] = hashedPassword
			needsUpdate = true
		}
		if existingUser.Name != userToSeed.Name {
			utils.SLog.Infof("Sistem kullanıcısı '%s' adı güncellenmeli.", userToSeed.Account)
			updateFields["name"] = userToSeed.Name
			needsUpdate = true
		}
		if !existingUser.Status {
			utils.SLog.Infof("Sistem kullanıcısı '%s' durumu güncellenmeli (true olarak ayarlanıyor).", userToSeed.Account)
			updateFields["status"] = true
			needsUpdate = true
		}

		if needsUpdate {
			utils.SLog.Infof("Mevcut sistem kullanıcısı '%s' güncelleniyor...", userToSeed.Account)
			err = db.Model(&existingUser).Updates(updateFields).Error
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
	err = db.Create(&userToSeed).Error
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
