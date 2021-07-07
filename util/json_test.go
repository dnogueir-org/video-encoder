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

	isJson := util.IsJson(json)
	require.Equal(t, true, isJson)

	json = `leinad`
	isJson = util.IsJson(json)
	require.Equal(t, false, isJson)
}
