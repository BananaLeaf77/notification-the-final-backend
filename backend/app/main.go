package main

import (
	"notification/config"
	"os"
	"os/signal"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger
var wg sync.WaitGroup

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error loading .env file")
	}

	log = config.GetLogrusInstance()

	startHTTP()
}

func startHTTP() {
	log.Info("Starting HTTP")
	app := fiber.New()

	// CORS Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	db, err := config.BootDB()
	if err != nil {
		log.Fatal("Failed to boot DB")
		return
	}

	// Regis repo and Usecase Here

	// delivery here

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server for Public on port %s", config.GetFiberHttpPort())
		if err := app.Listen(config.GetFiberListenAddress()); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	<-signalChan

	log.Info("Shutting down the server...")

	if err := app.Shutdown(); err != nil {
		log.Errorf("Error during server shutdown: %v", err)
	}

	wg.Wait()
	log.Info("Server shut down gracefully")
}
