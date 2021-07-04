package services

import (
	"dnogueir-org/video-encoder/internal/models"
	"dnogueir-org/video-encoder/util"
	"encoding/json"
	"os"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type JobWorkerResult struct {
	Job     models.Job
	Message *amqp.Delivery
	Error   error
}

var Mutex = &sync.Mutex{}

func JobWorker(messageChannel chan amqp.Delivery, returnChannel chan JobWorkerResult, jobService JobService, job models.Job, workerID int) {

	for message := range messageChannel {

		err := util.IsJson(string(message.Body))
		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		Mutex.Unlock()

		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		err = jobService.VideoService.Video.Validate()
		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("outputBucketName")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		Mutex.Lock()
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		jobService.Job = &job
		err = jobService.Start()
		if err != nil {
			returnChannel <- returnJobResult(models.Job{}, message, err)
			continue
		}

		returnChannel <- returnJobResult(job, message, nil)
	}
}

func returnJobResult(job models.Job, message amqp.Delivery, err error) JobWorkerResult {
	result := JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err,
	}

	return result
}
