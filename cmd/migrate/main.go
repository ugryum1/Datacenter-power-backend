package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"datacenter-power/internal/app/ds"
	"datacenter-power/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.DatacenterEngineer{},
		&ds.DatacenterDevice{},
		&ds.DatacenterDeviceLike{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
