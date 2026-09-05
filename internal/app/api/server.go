package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/handler"
	"datacenter-power/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	deviceRepository, err := repository.NewDatacenterDeviceRepository()
	if err != nil {
		logrus.Fatal("ошибка инициализации репозитория устройств дата-центра: ", err)
	}

	deviceHandler := handler.NewDatacenterDeviceHandler(deviceRepository)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/dc-equipment", deviceHandler.ShowEquipmentGrid)
	r.GET("/dc-equipment/reel/:id", deviceHandler.ShowEquipmentReel)
	r.GET("/dc-equipment/draft", deviceHandler.ShowEquipmentDraft)

	if err := r.Run(); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
