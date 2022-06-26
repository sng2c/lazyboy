package queue

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"lazyboy/tmpl"
	"log"
	"net/http"
	"path"
	"text/template"
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
type Req struct {
	Method  string
	Url     string
	Headers [][]string
	Body    string
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

func (pipe *Pipeline) BuildRequest(data interface{}) (*http.Request, error) {
	reqTmpl, err := pipe.ReqTmpl()
	if err != nil {
		return nil, err
	}
	log.Println("!")
	resolveTemplateData, err := tmpl.ResolveTemplate(reqTmpl, data)
	log.Println("!!")
	log.Println(resolveTemplateData)

	if err != nil {
		return nil, err
	}
	var req Req
	err = json.Unmarshal([]byte(resolveTemplateData), &req)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(req.Body)
	httpReq, err := http.NewRequest(req.Method, req.Url, buf)
	if err != nil {
		return nil, err
	}

	for _, hdrPair := range req.Headers {
		httpReq.Header.Add(hdrPair[0], hdrPair[1])
	}

	return httpReq, nil
}

func (pipe *Pipeline) ParseResponse(response *http.Response) (interface{}, error) {
	resTmpl, err := pipe.ResTmpl()
	if err != nil {
		return nil, err
	}
	resolveTemplateData, err := tmpl.ResolveTemplate(resTmpl, response)
	if err != nil {
		return nil, err
	}
	var res interface{}
	err = json.Unmarshal([]byte(resolveTemplateData), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
