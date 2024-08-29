package main

import (
	"fmt"
	"notification/config"
	"notification/services/notification/delivery"
	"notification/services/notification/repository"
	"notification/services/notification/usecase"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v4"
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

	// INIT SERVICES
	mClient, schoolPhone, err := config.InitMailgun()
	if err != nil{
		fmt.Println(err)
		panic(err)
	}

	// Register repository and Usecase here
	studentParentRepo := repository.NewStudentParentRepository(db)
	studentParentUC := usecase.NewStudentParentUseCase(studentParentRepo, 30*time.Second)

	mailGunRepo := repository.NewMailgunRepository(mClient, *schoolPhone)
	mailgunUC := usecase.NewMailGunUseCase(mailGunRepo, 30*time.Second)

	// Register delivery here
	delivery.NewStudentParentHandler(app, studentParentUC)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server for Public on port %s", config.GetFiberHttpPort())
		if err := app.Listen(config.GetFiberListenAddress()); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan

	log.Info("Shutting down the server...")

	if err := app.Shutdown(); err != nil {
		log.Errorf("Error during server shutdown: %v", err)
	}

	wg.Wait()
	log.Info("Server shut down gracefully")
}
