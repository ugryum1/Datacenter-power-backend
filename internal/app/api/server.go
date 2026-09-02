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

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatal("ошибка инициализации репозитория: ", err)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/devices", h.GetTile)
	r.GET("/feed/:id", h.GetFeed)
	r.GET("/add", h.GetAdd)

	if err := r.Run(); err != nil {
		logrus.Error(err)
	}

	log.Println("Server down")
}
