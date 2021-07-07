package database

import (
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
)

type Database struct {
	db            *gorm.DB
	dsn           string
	dbType        string
	debug         bool
	autoMigrateDb bool
	env           string
}

func NewDb(dsn string, dbType string, debug bool, autoMigrateDb bool, env string) *Database {
	return &Database{
		dsn:           dsn,
		dbType:        dbType,
		debug:         debug,
		autoMigrateDb: autoMigrateDb,
		env:           env,
	}
}

func NewDbTest() *gorm.DB {

	dbInstance := NewDb(":memory:", "sqlite3", true, true, "test")
	connection, err := dbInstance.Connect()

	if err != nil {
		internal.Logger.Fatal(err.Error())
	}

	return connection
}

func (d *Database) Connect() (*gorm.DB, error) {

	var err error

	d.db, err = gorm.Open(d.dbType, d.dsn)

	if err != nil {
		return nil, err
	}

	d.db.LogMode(d.debug)

	if d.autoMigrateDb {
		d.db.AutoMigrate(&models.Video{}, &models.Job{})
		d.db.Model(models.Job{}).AddForeignKey("video_id", "videos (id)", "CASCADE", "CASCADE")
	}

	return d.db, nil
}
