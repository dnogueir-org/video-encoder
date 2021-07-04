package repository

import (
	"dnogueir-org/video-encoder/internal/models"
	"fmt"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type VideoRepository interface {
	Insert(video *models.Video) (*models.Video, error)
	Find(id string) (*models.Video, error)
}

type VideoRepositoryDb struct {
	Db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepositoryDb {
	return &VideoRepositoryDb{Db: db}
}

func (repo VideoRepositoryDb) Insert(video *models.Video) (*models.Video, error) {

	if video.ID == "" {
		video.ID = uuid.NewV4().String()
	}

	err := repo.Db.Create(video).Error

	if err != nil {
		return nil, err
	}

	return video, nil
}

func (repo VideoRepositoryDb) Find(id string) (*models.Video, error) {

	var video models.Video
	repo.Db.Preload("Jobs").First(&video, "id = ?", id)

	if video.ID == "" {
		return nil, fmt.Errorf("video does not exist")
	}

	return &video, nil

}
