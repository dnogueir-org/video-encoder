package services

import (
	"dnogueir-org/video-encoder/application/repositories"
	"dnogueir-org/video-encoder/domain"
	"dnogueir-org/video-encoder/framework/queue"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db               *gorm.DB
	Domain           domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManager(db *gorm.DB, rabbitMQ *queue.RabbitMQ, jobReturnChannel chan JobWorkerResult, messageChannel chan amqp.Delivery) *JobManager {
	return &JobManager{
		Db:               db,
		Domain:           domain.Job{},
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (jm *JobManager) Start(ch *amqp.Channel) {

	videoService := NewVideoService()
	videoService.VideoRepository = repositories.VideoRepositoryDb{Db: jm.Db}

	jobService := JobService{
		JobRepository: repositories.JobRepositoryDb{Db: jm.Db},
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))
	if err != nil {
		log.Fatalf("error loading var: CONCURRENCY_WORKERS")
	}

	for processQuantity := 0; processQuantity < concurrency; processQuantity++ {
		go JobWorker(jm.MessageChannel, jm.JobReturnChannel, jobService, jm.Domain, processQuantity)
	}

	for jobResult := range jm.JobReturnChannel {
		if jobResult.Error != nil {
			err = jm.checkParseErrors(jobResult)
		} else {
			err = jm.notifySuccess(jobResult, ch)
		}

		if err != nil {
			jobResult.Message.Reject(false)
		}
	}

}

func (jm *JobManager) notifySuccess(jobResult JobWorkerResult, ch *amqp.Channel) error {

	jobJson, err := json.Marshal(jobResult.Job)

	if err != nil {
		return err
	}

	err = jm.notify(jobJson)

	if err != nil {
		return err
	}

	err = jobResult.Message.Ack(false)
	if err != nil {
		return err
	}

	return nil
}

func (jm *JobManager) checkParseErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Printf("messageID #{jobResult.Message.DeliveryTag}. Error passing job #{jobResult.Job.ID}")
	} else {
		log.Printf("messageID #{jobResult.Message.DeliveryTag}. Error parsing message #{jobResult.Job.Error}")
	}

	errorMessage := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error:   jobResult.Error.Error(),
	}

	jobJson, err := json.Marshal(errorMessage)

	err = jm.notify(jobJson)
	if err != nil {
		return err
	}

	err = jobResult.Message.Reject(false)

	if err != nil {
		return err
	}

	return nil

}

func (jm JobManager) notify(jobJson []byte) error {

	err := jm.RabbitMQ.Notify(
		string(jobJson),
		"application/json",
		os.Getenv("RABBITMQ_NOTIFICATION_EX"),
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"),
	)

	if err != nil {
		return err
	}

	return nil
}
