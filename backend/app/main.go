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
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger
var wg sync.WaitGroup

func main() {
	log = config.GetLogrusInstance()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	startHTTP()
}

func startHTTP() {
	log.Info("Starting HTTP")
	app := fiber.New(config.GetFiberConfig())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("OK")
	})

	db, err := config.BootDB()
	if err != nil {
		log.Fatal("Failed to boot DB")
		return
	}

	meow, eAuth, eAdress, schoolPhone, emailSender, err := config.InitSender()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Failed to boot Sender Service")
		return
	}

	// Repo And Usecase Declare mafaka
	// Notification
	notifRepo := repository.NewNotificationRepository(db)
	notifUC := usecase.NewNotificationUseCase(notifRepo)
	// Auth
	authRepo := repository.NewAuthRepository(db)
	authUC := usecase.NewAuthUseCase(authRepo)
	// User
	userRepo := repository.NewUserRepository(db)
	userUC := usecase.NewUserUseCase(userRepo, 100*time.Second)
	// StudentParent
	studentParentRepo := repository.NewStudentParentRepository(db)
	studentParentUC := usecase.NewStudentParentUseCase(studentParentRepo, 30*time.Second)
	// Student
	studentRepo := repository.NewStudentRepository(db)
	studentUC := usecase.NewStudentUseCase(studentRepo, 100*time.Second)
	// Sender
	senderRepo := repository.NewSenderRepository(db, eAuth, *eAdress, *schoolPhone, *emailSender, meow)
	senderUC := usecase.NewSenderUseCase(senderRepo, 30*time.Second)
	// Register delivery here
	delivery.NewNotificationHandler(app, notifUC)
	delivery.NewUserAuthHandler(app, authUC)
	delivery.NewUserHandler(app, userUC)
	delivery.NewStudentParentHandler(app, studentParentUC)
	delivery.NewSenderDelivery(app, senderUC)
	delivery.NewStudentDelivery(app, studentUC)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %s", config.GetFiberHttpPort())
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
