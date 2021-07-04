package repository_test

import (
	"dnogueir-org/video-encoder/database"
	"dnogueir-org/video-encoder/models"
	"dnogueir-org/video-encoder/repository"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestVideoRepositoryDbInsert(t *testing.T) {
	db := database.NewDbTest()
	defer db.Close()

	video := models.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repository.VideoRepositoryDb{Db: db}
	repo.Insert(video)

	v, err := repo.Find(video.ID)

	require.NotEmpty(t, v.ID)
	require.Nil(t, err)
	require.Equal(t, v.ID, video.ID)
}
