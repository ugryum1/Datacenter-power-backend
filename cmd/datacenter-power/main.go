package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"datacenter-power/internal/app/config"
	"datacenter-power/internal/app/dsn"
	"datacenter-power/internal/app/handler"
	"datacenter-power/internal/app/repository"
	"datacenter-power/internal/pkg"
)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
