package utils_test

import (
	"dnogueir-org/video-encoder/framework/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJson(t *testing.T) {
	json := `{
		"id": "1234",
		"file_path": "bachianinha.mp4",
		"status": "downloading"
	}`

	err := utils.IsJson(json)
	require.Nil(t, err)

	json = `leinad`
	err = utils.IsJson(json)
	require.NotNil(t, err)
}
