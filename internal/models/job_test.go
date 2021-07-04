package models_test

import (
	"dnogueir-org/video-encoder/internal/models"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestNewJob(t *testing.T) {
	video := models.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	job, err := models.NewJob("path", "Converted", video)
	require.NotNil(t, job)
	require.Nil(t, err)
}
