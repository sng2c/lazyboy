package proc

import (
	"bytes"
	"encoding/json"
	"lazyboy/queue"
	"net/http"
)

func BuildHttpRequest(req *queue.Req) (*http.Request, error) {
	var bodyBuf *bytes.Buffer

	switch req.BodyType {
	case queue.BodyTypeJson:
		jsonStr, err := json.Marshal(req.BodyObj)
		if err != nil {
			return nil, err
		}
		bodyBuf = bytes.NewBuffer(jsonStr)
	case queue.BodyTypeText:
		bodyBuf = bytes.NewBufferString(req.BodyStr)
	case queue.BodyTypeByte:
		bodyBuf = bytes.NewBuffer(req.BodyBytes)
	default:
		bodyBuf = nil
	}

	request, err := http.NewRequest(req.Method, req.Url, bodyBuf)
	if err != nil {
		return nil, err
	}
	for _, headerPair := range req.Headers {
		request.Header.Add(headerPair[0], headerPair[1])
	}

	return request, nil
}
