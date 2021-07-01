package services

import (
	"dnogueir-org/video-encoder/application/repositories"
	"dnogueir-org/video-encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (js *JobService) Start() error {

	err := js.changeJobStatus("DOWNLOADING")
	if err != nil {
		return js.failJob(err)
	}

	err = js.VideoService.Download(os.Getenv("inputBucketName"))
	if err != nil {
		return js.failJob(err)
	}

	err = js.changeJobStatus("FRAGMENTING")
	if err != nil {
		return js.failJob(err)
	}

	err = js.VideoService.Fragment()
	if err != nil {
		return js.failJob(err)
	}

	err = js.changeJobStatus("ENCODING")
	if err != nil {
		return js.failJob(err)
	}

	err = js.VideoService.Encode()
	if err != nil {
		return js.failJob(err)
	}

	err = js.performUpload()
	if err != nil {
		return js.failJob(err)
	}

	err = js.changeJobStatus("FINISHING")
	if err != nil {
		return js.failJob(err)
	}

	err = js.VideoService.Finish()
	if err != nil {
		return js.failJob(err)
	}

	err = js.changeJobStatus("COMPLETED")
	if err != nil {
		return js.failJob(err)
	}

	return nil
}

func (js *JobService) performUpload() error {

	err := js.changeJobStatus("UPLOADING")
	if err != nil {
		return js.failJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("outputBucketName")
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + js.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))
	doneUpload := make(chan string)

	go videoUpload.ProcessUpload(concurrency, doneUpload)

	var uploadResult string
	uploadResult = <-doneUpload

	if uploadResult != "upload completed" {
		return js.failJob(errors.New(uploadResult))
	}

	return err

}

func (js *JobService) changeJobStatus(status string) error {

	var err error

	js.Job.Status = status
	js.Job, err = js.JobRepository.Update(js.Job)

	if err != nil {
		return js.failJob(err)
	}

	return nil
}

func (js *JobService) failJob(error error) error {

	js.Job.Status = "FAILED"
	js.Job.Error = error.Error()

	_, err := js.JobRepository.Update(js.Job)

	if err != nil {
		return err
	}

	return error
}
