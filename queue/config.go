package queue

import (
	"encoding/json"
	"net/http"
)

type QueueConfig struct {
	TakePerTick      uint64
	Correction       float64
	RequestTemplate  string
	ResponseTemplate string
}

type HttpRequest struct {
}
type HttpResponse struct {
}

func NewQueueConfig(jsonStr []byte) *QueueConfig {
	var conf QueueConfig
	err := json.Unmarshal(jsonStr, &conf)
	if err != nil {
		return nil
	}
	return &conf
}

func (conf *QueueConfig) BuildRequest(request HttpRequest) (*http.Request, error) {
	return nil, nil
}

func (conf *QueueConfig) ParseResponse(response http.Response) (HttpResponse, error) {
	return HttpResponse{}, nil
}
