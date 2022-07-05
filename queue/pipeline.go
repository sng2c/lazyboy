package queue

import (
	"encoding/json"
	"errors"
	"github.com/PaesslerAG/jsonpath"
	"github.com/robfig/cron"
	logrus "github.com/sirupsen/logrus"
	"io/ioutil"
	"lazyboy/tmpl"
	"path"
	"text/template"
	"time"
)

type Pipeline struct {
	Name          string
	UniqueKey     string
	TakePerTick   int
	ActiveTime    string
	Workers       int
	ReqTmplName   string
	ResTmplName   string
	ResBodyType   BodyType
	OutputPath    string
	reqTmplString string
	resTmplString string
	queuePath     string
}

func (pipe *Pipeline) OutputAbsPath() string {
	return path.Join(pipe.queuePath, pipe.OutputPath)
}
func (pipe *Pipeline) GetName() string {
	if pipe.Name == "" {
		return path.Base(pipe.queuePath)
	}
	return pipe.Name
}

func (pipe *Pipeline) GetUniqueKey(o interface{}) (interface{}, error) {
	return jsonpath.Get(pipe.UniqueKey, o)
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
	var pipe Pipeline
	err := json.Unmarshal(jsonStr, &pipe)
	if err != nil {
		return nil, err
	}

	pipe.queuePath = basePath

	// load req
	if pipe.ReqTmplName != "" {
		b, err := ioutil.ReadFile(path.Join(pipe.queuePath, pipe.ReqTmplName))
		if err != nil {
			return nil, err
		}
		pipe.reqTmplString = string(b)
	}
	// load res
	if pipe.ResTmplName != "" {
		b, err := ioutil.ReadFile(path.Join(pipe.queuePath, pipe.ResTmplName))
		if err != nil {
			return nil, err
		}
		pipe.resTmplString = string(b)
	}

	if pipe.OutputPath == "" {
		return nil, errors.New("OutputPath is required")
	}

	if pipe.UniqueKey == "" {
		return nil, errors.New("UniqueKey is required")
	}

	return &pipe, nil
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
	return pipe.TakePerTick
}
func (pipe *Pipeline) Take() [][]byte {
	var gTaken = make([][]byte, 0)
	var want = pipe.WantToTake()
	for want > 0 {
		queue, err := OfferFileQueue(pipe.queuePath)
		if err != nil {
			logrus.WithFields(logrus.Fields{"ctx": "queue/Pipeline.Take", "path": pipe.queuePath}).Debug("no more data.", err)
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
