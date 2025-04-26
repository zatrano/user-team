package main

import (
	"flag"

	"zatrano/configs"
	"zatrano/database"
	"zatrano/utils"
)

func main() {
	utils.InitLogger()
	defer utils.SyncLogger()
	migrateFlag := flag.Bool("migrate", false, "Veritabanı başlatma işlemini çalıştır (migrasyonları içerir)")
	seedFlag := flag.Bool("seed", false, "Veritabanı başlatma işlemini çalıştır (seederları içerir)")
	flag.Parse()

	configs.InitDB()
	defer configs.CloseDB()

	db := configs.GetDB()

	utils.SLog.Info("Veritabanı başlatma işlemi çalıştırılıyor...")
	database.Initialize(db, *migrateFlag, *seedFlag)

	utils.SLog.Info("Veritabanı başlatma işlemi tamamlandı.")
}
