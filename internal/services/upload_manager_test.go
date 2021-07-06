package services_test

import (
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/services"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		internal.Logger.Fatal("Error loading .env file")
	}
}

func TestVideoServiceUpload(t *testing.T) {
	video, repo := prepare()
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	bucketName := "video-encoder-bucket"
	err := videoService.Download(bucketName)
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = "video-encoder-bucket"
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(20, doneUpload)

	result := <-doneUpload
	require.Equal(t, "upload completed", result)

	err = videoService.Finish()
	require.Nil(t, err)
}
