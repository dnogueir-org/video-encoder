package main

import (
	"dnogueir-org/video-encoder/database"
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/services"
	"dnogueir-org/video-encoder/queue"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var db *database.Database

func init() {
	err := godotenv.Load()
	if err != nil {
		internal.Logger.Fatal("error loading .env file")
	}

	autoMmigrateDb, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		internal.Logger.Fatal("error parsing boolean env var")
	}

	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		internal.Logger.Fatal("error parsing boolean env var")
	}

	db = database.NewDb(
		os.Getenv("DSN"),
		os.Getenv("DB_TYPE"),
		debug,
		autoMmigrateDb,
		os.Getenv("ENV"))
}

func main() {

	messageChannel := make(chan amqp.Delivery, 10)
	jobReturnChannel := make(chan services.JobWorkerResult, 10)

	dbConnection, err := db.Connect()
	if err != nil {
		internal.Logger.Fatal("error connecting to DB")
	}

	defer dbConnection.Close()

	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()

	rabbitMQ.Consume(messageChannel)

	jobManager := services.NewJobManager(dbConnection, rabbitMQ, jobReturnChannel, messageChannel)
	jobManager.Start()
}
