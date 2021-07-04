package util_test

import (
	"dnogueir-org/video-encoder/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `{
		"id": "1234",
		"file_path": "bachianinha.mp4",
		"status": "downloading"
	}`

	err := util.IsJson(json)
	require.Nil(t, err)

	json = `leinad`
	err = util.IsJson(json)
	require.NotNil(t, err)
}
