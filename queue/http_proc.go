package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/fatih/structs"
	log "github.com/sirupsen/logrus"
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
	BodyObj   interface{}
	BodyBytes []byte
}

type Res struct {
	Req        *Req
	Status     string
	StatusCode int
	Headers    map[string]interface{}
	BodyType   BodyType
	BodyText   string
	BodyObj    interface{}
	BodyBytes  []byte
	Err        string
}

func NewReqFromPipeline(pipe *Pipeline, data interface{}) (*Req, error) {
	reqTmpl, err := pipe.ReqTmpl()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resolveTemplateData, err := tmpl.ResolveTemplate(reqTmpl, data)

	if err != nil {
		return nil, err
	}
	var req Req
	err = json.Unmarshal([]byte(resolveTemplateData), &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func NewResFromHttpResponse(response *http.Response) (*Res, error) {
	res := Res{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    map[string]interface{}{},
	}
	var err error
	if response.Body != nil {
		res.BodyBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}
	log.Println(response.Header.Get("Content-type"))
	conttype := response.Header.Get("Content-type")
	switch {
	case conttype == "application/json":
		res.BodyType = BodyTypeJson
		err := json.Unmarshal(res.BodyBytes, &res.BodyObj)
		if err != nil {
			return nil, err
		}
	case strings.HasPrefix(conttype, "text/"):
		res.BodyType = BodyTypeText
		if res.BodyBytes != nil {
			res.BodyText = string(res.BodyBytes)
		}
	default:
		res.BodyType = BodyTypeByte
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
	request, err := req.BuildHttpRequest()
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

	res, err = NewResFromHttpResponse(response)
	if err != nil {
		res = &Res{}
		res.Err = err.Error()
		res.Req = req
	}

	res.Req = req
	return res
}

func (res *Res) ParseResponse(pipe *Pipeline) ([]byte, error) {
	resTmpl, err := pipe.ResTmpl()
	if err != nil {
		return nil, err
	}
	return BuildResTemplate(resTmpl, res)
}

func BuildResTemplate(resTmpl *template.Template, res *Res) ([]byte, error) {
	log.Println("BuildResTemplate", res)
	resolveTemplateData, err := tmpl.ResolveTemplate(resTmpl, structs.Map(res))
	if err != nil {
		log.Println("BuildResTemplate err1", err)
		return nil, err
	}

	return resolveTemplateData, nil
}

func (req *Req) BuildHttpRequest() (*http.Request, error) {
	var bodyBuf *bytes.Buffer

	switch req.BodyType {
	case BodyTypeJson:
		jsonStr, err := json.Marshal(req.BodyObj)
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

	request, err := http.NewRequest(req.Method, req.Url, bodyBuf)
	if err != nil {
		return nil, err
	}
	for key, val := range req.Headers {
		request.Header.Add(key, val.(string))
	}

	return request, nil
}
