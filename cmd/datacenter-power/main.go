package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "datacenter-power/docs"
	"datacenter-power/internal/app/config"
	"datacenter-power/internal/app/dsn"
	"datacenter-power/internal/app/handler"
	"datacenter-power/internal/app/redis"
	"datacenter-power/internal/app/repository"
	"datacenter-power/internal/pkg"
)

//	@title			Datacenter Power API
//	@version		1.0
//	@description	Расчёт потребляемой датацентром электроэнергии по установленному оборудованию

//	@host		localhost:8080
//	@schemes	http
//	@BasePath	/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Значение заголовка: Bearer <access_token>

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()

	rep, errRep := repository.New(&repository.RepositorySettings{
		PostgresDSN:     postgresString,
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ROOT_USER"),
		MinioSecretKey:  os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioBucketName: os.Getenv("MINIO_BUCKET"),
	})
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(rep, redisClient, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
