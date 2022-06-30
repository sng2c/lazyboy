package queue

import (
	"encoding/json"
	"github.com/robfig/cron"
	"io/ioutil"
	"lazyboy/tmpl"
	"log"
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
