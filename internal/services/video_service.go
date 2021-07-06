package services

import (
	"context"
	"dnogueir-org/video-encoder/internal"
	"dnogueir-org/video-encoder/internal/models"
	"dnogueir-org/video-encoder/repository"
	"io/ioutil"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
)

type VideoService struct {
	Video           *models.Video
	VideoRepository repository.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(v.Video.FilePath)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f, err := os.Create(os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		return err
	}

	_, err = f.Write(body)
	if err != nil {
		return err
	}

	defer f.Close()

	internal.Logger.WithFields(logrus.Fields{
		"videoID": v.Video.ID,
	}).Info("video has been stored")

	return nil
}

func (v *VideoService) Fragment() error {
	err := os.Mkdir(os.Getenv("localStoragePath")+"/"+v.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}

	source := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {

	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv("localStoragePath")+"/"+v.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, os.Getenv("localStoragePath")+"/"+v.Video.ID)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")

	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil

}

func (v *VideoService) Finish() error {
	err := os.Remove(os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		logVideoError(err, v.Video.ID)
		return err
	}

	err = os.Remove(os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag")
	if err != nil {
		logVideoError(err, v.Video.ID)
		return err
	}

	err = os.RemoveAll(os.Getenv("localStoragePath") + "/" + v.Video.ID)
	if err != nil {
		logVideoError(err, v.Video.ID)
		return err
	}

	internal.Logger.Info("files have been removed")

	return nil

}

func (vs *VideoService) InsertVideo() error {
	_, err := vs.VideoRepository.Insert(vs.Video)
	if err != nil {
		return err
	}

	return nil
}

func logVideoError(err error, videoId string) {
	internal.Logger.WithFields(logrus.Fields{
		"videoId": videoId,
	}).Error(err.Error())
}

func printOutput(out []byte) {
	if len(out) > 0 {
		internal.Logger.Info(string(out))
	}
}
