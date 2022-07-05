package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/fatih/structs"
	logrus "github.com/sirupsen/logrus"
	"io/ioutil"
	"lazyboy/tmpl"
	"net/http"
	"strings"
	"text/template"
)

type BodyType string

const BodyTypeNone = BodyType("")
const BodyTypeText = BodyType("TEXT")
const BodyTypeJson = BodyType("JSON")
const BodyTypeByte = BodyType("BYTE")

type Req struct {
	Method    string
	Url       string
	Headers   map[string]interface{}
	Extra     map[string]interface{}
	BodyType  BodyType
	BodyStr   string
	BodyJson  interface{}
	BodyBytes []byte
}

type Res struct {
	Req        *Req
	Status     string
	StatusCode int
	Headers    map[string]interface{}
	BodyType   BodyType
	BodyText   string
	BodyJson   interface{}
	BodyBytes  []byte
	Err        string
}

var ResTmplFormatError = errors.New("ResTmpl must be JSON format.")

func NewReqFromPipeline(pipe *Pipeline, data interface{}) (*Req, error) {
	logger := logrus.WithFields(logrus.Fields{"ctx": "http_proc/NewReqFromPipeline", "path": pipe.queuePath})
	reqTmpl, err := pipe.ReqTmpl()
	if err != nil {
		logger.Warn(err)
		return nil, err
	}

	resolveTemplateData, err := tmpl.ResolveTemplate(reqTmpl, data)

	if err != nil {
		logger.Warn(err)
		return nil, err
	}
	var req Req
	err = json.Unmarshal([]byte(resolveTemplateData), &req)
	if err != nil {
		logger.Warn(err)
		return nil, err
	}

	return &req, nil
}

func NewResFromHttpResponse(response *http.Response, forcedBodyType BodyType) (*Res, error) {
	logger := logrus.WithFields(logrus.Fields{"ctx": "http_proc/NewResFromHttpResponse"})
	res := Res{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    map[string]interface{}{},
	}
	var err error
	if response.Body != nil {
		res.BodyBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			logger.Warn(err)
			return nil, err
		}
	}
	logrus.Debugf("Content-type : %v", response.Header.Get("Content-type"))
	contType := response.Header.Get("Content-type")

	switch {
	case contType == "application/json":
		res.BodyType = BodyTypeJson
	case strings.HasPrefix(contType, "text/"):
		res.BodyType = BodyTypeText
	default:
		res.BodyType = BodyTypeByte
	}

	if forcedBodyType != BodyTypeNone {
		logrus.Debugf("Enforce bodyType to %v", forcedBodyType)
		res.BodyType = forcedBodyType
	}

	switch res.BodyType {
	case BodyTypeJson:
		err := json.Unmarshal(res.BodyBytes, &res.BodyJson)
		if err != nil {
			logger.Warn(err)
			return nil, err
		}
	case BodyTypeText:
		if res.BodyBytes != nil {
			res.BodyText = string(res.BodyBytes)
		}
	}

	for k, v := range response.Header {
		for _, vv := range v {
			res.Headers[k] = vv
		}
	}
	return &res, nil
}
func (req *Req) Run(ctx context.Context, pipe *Pipeline) *Res {
	var res *Res

	request, err := req.BuildHttpRequest(ctx)

	if err != nil {
		res = &Res{}
		res.Err = err.Error()
		res.Req = req
		return res
	}
	ua := &http.Client{}
	response, err := ua.Do(request)
	if err != nil {
		res = &Res{}
		res.Err = err.Error()
		res.Req = req
		return res
	}

	res, err = NewResFromHttpResponse(response, pipe.ResBodyType)
	if err != nil {
		res = &Res{}
		res.Err = err.Error()
		res.Req = req
	}

	res.Req = req
	return res
}

func (res *Res) BuildOutput(pipe *Pipeline) (interface{}, error) {
	resTmpl, err := pipe.ResTmpl()
	if err != nil {
		return nil, err
	}
	resTemplate, err := BuildResTemplate(resTmpl, res)
	if err != nil {
		return nil, err
	}
	var out interface{}
	err = json.Unmarshal(resTemplate, &out)
	if err != nil {
		logrus.Warn("ResTmpl must be JSON format.")
		return nil, ResTmplFormatError
	}
	return out, nil
}

func BuildResTemplate(resTmpl *template.Template, res *Res) ([]byte, error) {
	logrus.Debug("BuildResTemplate", res)
	resolveTemplateData, err := tmpl.ResolveTemplate(resTmpl, structs.Map(res))
	if err != nil {
		logrus.Debug("BuildResTemplate err1", err)
		return nil, err
	}

	return resolveTemplateData, nil
}

func (req *Req) BuildHttpRequest(ctx context.Context) (*http.Request, error) {
	var bodyBuf *bytes.Buffer

	switch req.BodyType {
	case BodyTypeJson:
		jsonStr, err := json.Marshal(req.BodyJson)
		if err != nil {
			return nil, err
		}
		bodyBuf = bytes.NewBuffer(jsonStr)
	case BodyTypeText:
		bodyBuf = bytes.NewBufferString(req.BodyStr)
	case BodyTypeByte:
		bodyBuf = bytes.NewBuffer(req.BodyBytes)
	default:
		bodyBuf = bytes.NewBuffer([]byte{})
	}

	request, err := http.NewRequestWithContext(ctx, req.Method, req.Url, bodyBuf)
	if err != nil {
		return nil, err
	}
	for key, val := range req.Headers {
		request.Header.Add(key, val.(string))
	}

	return request, nil
}
