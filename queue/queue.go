package queue

import (
	"encoding/json"
	"io/ioutil"
	"lazyboy/tmpl"
	"log"
	"net/http"
	"path"
	"text/template"
)

type Queue struct {
	TakePerTick      uint64
	Correction       float64
	QueuePath        string `json:"-"`
	ReqTmplFilename  string
	ResTmplFilename  string
	ReqTmplString    string
	ResTmplString    string
	requestTemplate  *template.Template
	responseTemplate *template.Template
}

func NewQueueFromConfigPath(configPath string) (*Queue, error) {
	basePath := path.Dir(configPath)
	jsonStr, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return NewQueue(basePath, jsonStr)
}

func NewQueue(basePath string, jsonStr []byte) (*Queue, error) {
	var q Queue
	err := json.Unmarshal(jsonStr, &q)
	if err != nil {
		return nil, err
	}

	q.QueuePath = basePath

	// load req
	if q.ReqTmplFilename != "" {
		b, err := ioutil.ReadFile(path.Join(q.QueuePath, q.ReqTmplFilename))
		if err != nil {
			return nil, err
		}
		q.ReqTmplString = string(b)
	}
	// load res
	if q.ResTmplFilename != "" {
		b, err := ioutil.ReadFile(path.Join(q.QueuePath, q.ResTmplFilename))
		if err != nil {
			return nil, err
		}
		q.ResTmplString = string(b)
	}

	return &q, nil
}
func (q *Queue) Init() error {
	var err error
	q.requestTemplate, err = tmpl.NewTemplate(q.ReqTmplString)
	if err != nil {
		return err
	}
	q.responseTemplate, err = tmpl.NewTemplate(q.ResTmplString)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) Take(n int) [][]byte {
	var gTaken [][]byte
	left := n
	for left > 0 {
		queue, err := OfferSubQueue(q.QueuePath)
		if err != nil {
			log.Println(err)
			break
		}

		taken := queue.Take(left)
		if taken != nil {
			for _, b := range taken {
				gTaken = append(gTaken, b)
			}
		}
	}
	return gTaken
}

func (q *Queue) BuildRequest(data interface{}) (*http.Request, error) {
	resolveTemplateData, err := tmpl.ResolveTemplate(q.requestTemplate, data)
	if err != nil {
		return nil, err
	}
	var req http.Request
	err = json.Unmarshal([]byte(resolveTemplateData), &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (q *Queue) ParseResponse(response *http.Response) (interface{}, error) {
	resolveTemplateData, err := tmpl.ResolveTemplate(q.responseTemplate, response)
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
