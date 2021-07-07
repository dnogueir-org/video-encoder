package util

import (
	"dnogueir-org/video-encoder/internal"
	"encoding/json"
)

func IsJson(s string) bool {
	var js struct{}

	err := json.Unmarshal([]byte(s), &js)

	if err != nil {
		internal.Logger.WithField("field", s).Error("This is not a json")
		return false
	}

	return true
}
