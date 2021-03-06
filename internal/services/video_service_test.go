package services_test

import (
	"dnogueir-org/video-encoder/database"
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/models"
	"dnogueir-org/video-encoder/internal/services"
	"dnogueir-org/video-encoder/repository"
	"testing"
	"time"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		internal.Logger.Error("Error loading .env file")
	}
}

func prepare() (*models.Video, repository.VideoRepositoryDb) {
	db := database.NewDbTest()
	defer db.Close()

	video := models.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "bachianinha.mp4"
	video.CreatedAt = time.Now()

	repo := repository.VideoRepositoryDb{Db: db}
	repo.Insert(video)

	return video, repo
}

func TestVideoServiceDownload(t *testing.T) {
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

	err = videoService.Finish()
	require.Nil(t, err)
}
