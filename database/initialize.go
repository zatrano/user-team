package database

import (
	"zatrano/database/migrations"
	"zatrano/database/seeders"
	"zatrano/models"
	"zatrano/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Initialize(db *gorm.DB, migrate bool, seed bool) {
	if !migrate && !seed {
		utils.SLog.Info("Migrate veya seed bayrağı belirtilmedi, işlem yapılmayacak.")
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.Log.Fatal("Veritabanı başlatma işlemi başarısız oldu, geri alındı (panic)", zap.Any("panic_info", r))
		}
		if tx.Error != nil && tx.Error != gorm.ErrInvalidTransaction {
			utils.SLog.Warn("Başlatma sırasında hata oluştuğu için işlem geri alınıyor.")
			tx.Rollback()
		}
	}()

	utils.SLog.Info("Veritabanı başlatma işlemi başlıyor...")

	if migrate {
		utils.SLog.Info("Migrasyonlar çalıştırılıyor...")
		if err := RunMigrationsInOrder(tx); err != nil {
			tx.Rollback()
			utils.Log.Fatal("Migrasyon başarısız oldu", zap.Error(err))
		}
		utils.SLog.Info("Migrasyonlar tamamlandı.")
	} else {
		utils.SLog.Info("Migrate bayrağı belirtilmedi, migrasyon adımı atlanıyor.")
	}

	if seed {
		utils.SLog.Info("Seeder'lar çalıştırılıyor...")
		if err := CheckAndRunSeeders(tx); err != nil {
			tx.Rollback()
			utils.Log.Fatal("Seeding başarısız oldu", zap.Error(err))
		}
		utils.SLog.Info("Seeder'lar tamamlandı.")
	} else {
		utils.SLog.Info("Seed bayrağı belirtilmedi, seeder adımı atlanıyor.")
	}

	utils.SLog.Info("İşlem commit ediliyor...")
	if err := tx.Commit().Error; err != nil {
		utils.Log.Fatal("Commit başarısız oldu", zap.Error(err))
	}

	utils.SLog.Info("Veritabanı başlatma işlemi başarıyla tamamlandı")
}

func RunMigrationsInOrder(db *gorm.DB) error {
	utils.SLog.Info(" -> Team migrasyonları çalıştırılıyor...")
	if err := migrations.MigrateTeamsTable(db); err != nil {
		utils.Log.Error("Teams tablosu migrasyonu başarısız oldu", zap.Error(err))
		return err
	}
	utils.SLog.Info(" -> Team migrasyonları tamamlandı.")

	utils.SLog.Info(" -> User migrasyonları çalıştırılıyor...")
	if err := migrations.MigrateUsersTable(db); err != nil {
		utils.Log.Error("Users tablosu migrasyonu başarısız oldu", zap.Error(err))
		return err
	}
	utils.SLog.Info(" -> User migrasyonları tamamlandı.")

	utils.SLog.Info("Tüm migrasyonlar başarıyla çalıştırıldı.")
	return nil
}

func CheckAndRunSeeders(db *gorm.DB) error {
	systemUser := seeders.GetSystemUserConfig()
	var existingUser models.User
	result := db.Where("account = ? AND type = ?", systemUser.Account, models.System).First(&existingUser)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.SLog.Infof("Sistem kullanıcısı oluşturuluyor: %s (%s)...", systemUser.Name, systemUser.Account)
			if err := seeders.SeedSystemUser(db); err != nil {
				utils.Log.Error("Sistem kullanıcısı seed edilemedi", zap.Error(err))
				return err
			}
			utils.SLog.Info(" -> Sistem kullanıcısı oluşturuldu.")
		} else {
			utils.Log.Error("Sistem kullanıcısı kontrol edilirken hata", zap.Error(result.Error))
			return result.Error
		}
	} else {
		utils.SLog.Infof("Sistem kullanıcısı '%s' (%s) zaten mevcut, oluşturma adımı atlanıyor.",
			existingUser.Name, existingUser.Account)
		utils.SLog.Infof("Mevcut sistem kullanıcısı '%s' için güncelleme kontrolü yapılıyor...", existingUser.Account)
		if err := seeders.SeedSystemUser(db); err != nil {
			utils.Log.Error("Mevcut sistem kullanıcısı güncellenirken/kontrol edilirken hata", zap.Error(err))
			return err
		}

	}
	return nil
}
