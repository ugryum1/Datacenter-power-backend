package main

import (
	"log"

	"datacenter-power/internal/app/api"
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
