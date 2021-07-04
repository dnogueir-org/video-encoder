package repository

import (
	"dnogueir-org/video-encoder/internal/models"
	"fmt"

	"github.com/jinzhu/gorm"
)

type JobRepository interface {
	Insert(job *models.Job) (*models.Job, error)
	Find(id string) (*models.Job, error)
	Update(job *models.Job) (*models.Job, error)
}

type JobRepositoryDb struct {
	Db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepositoryDb {
	return &JobRepositoryDb{Db: db}
}

func (repo JobRepositoryDb) Insert(job *models.Job) (*models.Job, error) {

	err := repo.Db.Create(job).Error

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (repo JobRepositoryDb) Find(id string) (*models.Job, error) {

	var job models.Job
	repo.Db.Preload("Video").First(&job, "id = ?", id)

	if job.ID == "" {
		return nil, fmt.Errorf("job does not exist")
	}

	return &job, nil

}

func (repo JobRepositoryDb) Update(job *models.Job) (*models.Job, error) {

	err := repo.Db.Save(&job).Error

	if err != nil {
		return nil, err
	}

	return job, nil
}
