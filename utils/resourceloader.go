package utils

import (
	logger "github.com/sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
)

func LoadConfig(fileName string, conf interface{}) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Error("[loadConfig]: %s\n", err)
	}
	err = json.Unmarshal(file, &conf)
	if err != nil {
		logger.Error("[loadConfig]: %s\n", err)
	}
}
