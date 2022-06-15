package queue

import (
	"encoding/json"
)

type QueueConfig struct {
	TakePerTick      uint64
	Correction       float64
	RequestTemplate  string
	ResponseTemplate string
}

func NewQueueConfig(jsonStr []byte) *QueueConfig {
	var conf QueueConfig
	err := json.Unmarshal(jsonStr, &conf)
	if err != nil {
		return nil
	}
	return &conf
}
