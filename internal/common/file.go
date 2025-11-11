package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func ReadDataToInterface[T any](data []byte, _ T) (*T, error) {

	var item T

	// remove all starting whitespace including newlines to figure out
	// what the first character is
	data = bytes.TrimLeftFunc(data, unicode.IsSpace)

	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")

	} else if len(data) < 10 {
		return nil, fmt.Errorf("data too short to determine format")

	} else if data[0] == '{' || data[0] == '[' {
		// If JSON we can unmarshal directly
		logrus.Debugln("Data format detected: JSON")
	} else {
		// If YAML we need to convert to JSON after. Have to use json
		// as the DSL serverless workflow SDK expects JSON
		var yamlData any
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			logrus.WithError(err).Errorln("Failed to unmarshal YAML")
			return nil, err
		}

		if jsonData, err := json.Marshal(yamlData); err != nil {
			logrus.WithError(err).Errorln("Failed to convert YAML to JSON")
			return nil, err
		} else {
			data = jsonData
		}
	}

	if err := json.Unmarshal(data, &item); err != nil {
		logrus.WithError(err).Errorln("Failed to unmarshal JSON data")
		return nil, fmt.Errorf("failed to unmarshal JSON file: %w", err)
	}

	return &item, nil
}
