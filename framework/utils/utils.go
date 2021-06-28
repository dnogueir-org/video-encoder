package utils

import "encoding/json"

func IsJson(s string) error {
	var js struct{}

	err := json.Unmarshal([]byte(s), &js)

	if err != nil {
		return err
	}

	return nil
}
