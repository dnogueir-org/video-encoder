package services

import (
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/models"
	"dnogueir-org/video-encoder/queue"
	"dnogueir-org/video-encoder/repository"
	"encoding/json"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db               *gorm.DB
	Model            models.Job
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
		Model:            models.Job{},
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (jm *JobManager) Start(ch *amqp.Channel) {

	videoService := NewVideoService()
	videoService.VideoRepository = repository.VideoRepositoryDb{Db: jm.Db}

	jobService := JobService{
		JobRepository: repository.JobRepositoryDb{Db: jm.Db},
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKER"))
	if err != nil {
		internal.Logger.Fatal("error loading var: CONCURRENCY_WORKER")
	}

	for processQuantity := 0; processQuantity < concurrency; processQuantity++ {
		go JobWorker(jm.MessageChannel, jm.JobReturnChannel, jobService, jm.Model, processQuantity)
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

	Mutex.Lock()
	jobJson, err := json.Marshal(jobResult.Job)
	Mutex.Unlock()

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
		internal.Logger.WithFields(logrus.Fields{
			"messageId": jobResult.Message.DeliveryTag,
			"jobId":     jobResult.Job.ID,
		}).Error(jobResult.Error.Error())
	} else {
		internal.Logger.WithFields(logrus.Fields{
			"messageId": jobResult.Message.DeliveryTag,
		}).Error(jobResult.Error.Error())
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
