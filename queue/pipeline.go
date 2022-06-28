package queue

import (
	"encoding/json"
	"github.com/fatih/structs"
	"github.com/robfig/cron"
	"io/ioutil"
	"lazyboy/tmpl"
	"log"
	"net/http"
	"path"
	"text/template"
	"time"
)

type Pipeline struct {
	TakePerTick   int
	Correction    float64
	ActiveTime    string
	ReqTmplName   string
	ResTmplName   string
	reqTmplString string `json:"-"`
	resTmplString string `json:"-"`
	queuePath     string
}
type BodyType string

const BodyTypeText = BodyType("Str")
const BodyTypeJson = BodyType("Json")
const BodyTypeByte = BodyType("Byte")

type Req struct {
	Method    string
	Url       string
	Headers   [][]string
	BodyType  BodyType
	BodyStr   string
	BodyObj   interface{}
	BodyBytes []byte
}

type Res struct {
	Req        *Req
	Status     string
	StatusCode int
	Headers    [][]string
	BodyType   BodyType
	BodyText   string
	BodyObj    interface{}
	BodyBytes  []byte
}

func NewPipelineFromConfigPath(configPath string) (*Pipeline, error) {
	basePath := path.Dir(configPath)
	jsonStr, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return NewPipeline(basePath, jsonStr)
}

func NewPipeline(basePath string, jsonStr []byte) (*Pipeline, error) {
	var q Pipeline
	err := json.Unmarshal(jsonStr, &q)
	if err != nil {
		return nil, err
	}

	q.queuePath = basePath

	// load req
	if q.ReqTmplName != "" {
		b, err := ioutil.ReadFile(path.Join(q.queuePath, q.ReqTmplName))
		if err != nil {
			return nil, err
		}
		q.reqTmplString = string(b)
	}
	// load res
	if q.ResTmplName != "" {
		b, err := ioutil.ReadFile(path.Join(q.queuePath, q.ResTmplName))
		if err != nil {
			return nil, err
		}
		q.resTmplString = string(b)
	}

	return &q, nil
}
func (pipe *Pipeline) IsActive(t time.Time) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow) // 분단위 cron
	spec, err := parser.Parse(pipe.ActiveTime)
	if err != nil {
		return false
	}
	t = t.Truncate(time.Minute)      // 분단위 밑으로는 제거
	tpast := t.Add(time.Minute * -1) // 1분전으로 기준시간 잡음
	next := spec.Next(tpast)
	return t.Equal(next)
}
func (pipe *Pipeline) ReqTmpl() (*template.Template, error) {
	var err error
	t, err := tmpl.NewTemplate(pipe.reqTmplString)
	if err != nil {
		return nil, err
	}
	return t, nil
}
func (pipe *Pipeline) ResTmpl() (*template.Template, error) {
	var err error
	t, err := tmpl.NewTemplate(pipe.resTmplString)
	if err != nil {
		return nil, err
	}
	return t, nil
}
func (pipe *Pipeline) WantToTake() int {
	return int(float64(pipe.TakePerTick) * pipe.Correction)
}
func (pipe *Pipeline) Take() [][]byte {
	var gTaken = make([][]byte, 0)
	var want = pipe.WantToTake()
	for want > 0 {
		queue, err := OfferFileQueue(pipe.queuePath)
		if err != nil {
			log.Println("no more data.", err)
			break
		}

		taken := queue.Take(want)
		if taken != nil {
			want -= len(taken)
			for _, b := range taken {
				gTaken = append(gTaken, b)
			}
		}
	}
	return gTaken
}

func (pipe *Pipeline) BuildReq(data interface{}) (*Req, error) {
	reqTmpl, err := pipe.ReqTmpl()
	if err != nil {
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

	//buf := bytes.NewBufferString(req.Body)
	//httpReq, err := http.NewRequest(req.Method, req.Url, buf)
	//if err != nil {
	//	return nil, err
	//}
	//
	//for _, hdrPair := range req.Headers {
	//	httpReq.Header.Add(hdrPair[0], hdrPair[1])
	//}

	return &req, nil
}

func (pipe *Pipeline) ParseResponse(response *http.Response) (interface{}, error) {
	resTmpl, err := pipe.ResTmpl()
	if err != nil {
		return nil, err
	}
	return parseResponse(resTmpl, response)
}

func NewResFromHttpResponse(response *http.Response) (*Res, error) {
	res := Res{
		Req:        nil,
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    [][]string{},
	}
	var err error
	if response.Body != nil {
		res.BodyBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	}
	log.Println(response.Header.Get("Content-type"))
	switch response.Header.Get("Content-type") {
	case "application/json":
		res.BodyType = BodyTypeJson
		err := json.Unmarshal(res.BodyBytes, &res.BodyObj)
		if err != nil {
			return nil, err
		}
	case "text/html":
		fallthrough
	case "text/plain":
		res.BodyType = BodyTypeText
		if res.BodyBytes != nil {
			res.BodyText = string(res.BodyBytes)
		}
	default:
		res.BodyType = BodyTypeByte
	}

	for k, v := range response.Header {
		for _, vv := range v {
			res.Headers = append(res.Headers, []string{k, vv})
		}
	}
	return &res, nil
}

func parseResponse(resTmpl *template.Template, response *http.Response) (interface{}, error) {
	res, err := NewResFromHttpResponse(response)
	if err != nil {
		return nil, err
	}
	log.Println(res)
	resolveTemplateData, err := tmpl.ResolveTemplate(resTmpl, structs.Map(res))
	if err != nil {
		return nil, err
	}
	var resobj interface{}
	log.Println(resolveTemplateData)
	err = json.Unmarshal([]byte(resolveTemplateData), &resobj)
	if err != nil {
		return nil, err
	}
	return resobj, nil
}
