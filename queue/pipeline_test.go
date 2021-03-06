package queue

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"lazyboy/tmpl"
	"path"
	"reflect"
	"testing"
	"text/template"
	"time"
)

func _convTemplate(str string) *template.Template {
	t, _ := tmpl.NewTemplate(str)
	return t
}
func TestNewQueueFromConfigPath(t *testing.T) {
	type args struct {
		configPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Pipeline
		wantErr bool
	}{
		{name: "pipeline_test1", args: args{configPath: path.Join(testBase, "pipeline_test1", "queue1", "_config.json")},
			want: &Pipeline{
				TakePerTick:   3,
				queuePath:     path.Join(testBase, "pipeline_test1", "queue1"),
				ReqTmplName:   "_req.json",
				ResTmplName:   "_res.json",
				reqTmplString: "{\n  \"name\": #{name}\n}",
				resTmplString: "{\n  \"body\" : #{$.body}\n}",
				OutputPath:    "./out.log",
				UniqueKey:     "$.uuid",
			}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPipelineFromConfigPath(tt.args.configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewPipelineFromConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPipelineFromConfigPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func newQueue(configPath string) *Pipeline {
	queue, err := NewPipelineFromConfigPath(configPath)
	if err != nil {
		log.Fatalln("NEW PIPE", err)
		return nil
	}
	return queue
}

func TestQueue_Take(t *testing.T) {
	type fields struct {
		Queue *Pipeline
	}
	type args struct {
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		{name: "take_test1", fields: fields{newQueue(path.Join(testBase, "pipeline_take_test1", "queue1", "_config.json"))},
			args: args{},
			want: [][]byte{
				[]byte("0"),
				[]byte("1"),
				[]byte("2"),
			}},
		{name: "take_test1", fields: fields{newQueue(path.Join(testBase, "pipeline_take_test1", "queue1", "_config.json"))},
			args: args{},
			want: [][]byte{
				[]byte("3"),
				[]byte("4"),
				[]byte("5"),
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fields.Queue
			if q == nil {
				t.Errorf("Pipeline failed construction")
			}
			if got := q.Take(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Take() = \n%v, \nwant \n%v", got, tt.want)
			}
		})
	}
}
func newPipeline(configPath string) *Pipeline {
	pipe, err := NewPipelineFromConfigPath(configPath)
	if err != nil {
		return nil
	}
	return pipe
}
func obj(s []byte) interface{} {
	var i interface{}
	err := json.Unmarshal(s, &i)
	if err != nil {
		return nil
	}
	if err != nil {
		log.Println(err)
		return nil
	}
	return i
}
func TestPipeline_BuildRequest(t *testing.T) {
	type fields struct {
		Pipeline *Pipeline
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Req
		wantErr bool
	}{
		{
			name:   "pipe buildReq",
			fields: fields{newQueue(path.Join(testBase, "pipelines", "pipe1", "config.json"))},
			args:   args{obj([]byte(`{"uuid":"1"}`))}, want: &Req{
			Method:  "GET",
			Url:     "http://localhost:8000/hello.txt?uuid=1",
			Headers: nil,
			Extra:   map[string]interface{}{"uuid": "1"},
		}, wantErr: false},
		{
			name:   "pipe buildReq2",
			fields: fields{newQueue(path.Join(testBase, "pipelines", "pipe2", "config.json"))},
			args:   args{obj([]byte(`{"uuid":"&1"}`))}, want: &Req{
			Method:  "POST",
			Url:     "http://localhost:8080/get",
			Headers: nil,
			BodyStr: `uuid=%261`,
		}, wantErr: false},
		{
			name:   "pipe buildReq2",
			fields: fields{newQueue(path.Join(testBase, "pipelines", "pipe2", "config.json"))},
			args:   args{obj([]byte(`{"uuid":"?????????"}`))}, want: &Req{
			Method:  "POST",
			Url:     "http://localhost:8080/get",
			Headers: nil,
			BodyStr: `uuid=%EA%B0%80%EB%82%98%EB%8B%A4`,
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipe := tt.fields.Pipeline

			got, err := NewReqFromPipeline(pipe, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildReq()\ngot = %#v,\nwant  %#v", got, tt.want)
			}
		})
	}
}

func TestPipeline_IsActive(t *testing.T) {
	type fields struct {
		ActiveTime string
	}
	type args struct {
		t time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "check all", fields: fields{ActiveTime: "* * * * *"}, args: args{time.Now()}, want: true},
		{name: "check specific", fields: fields{ActiveTime: "1 * * * *"}, args: args{time.Date(2022, 6, 28, 10, 0, 0, 0, time.Local)}, want: false},
		{name: "check specific", fields: fields{ActiveTime: "1 * * * *"}, args: args{time.Date(2022, 6, 28, 10, 1, 0, 0, time.Local)}, want: true},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 10, 1, 0, 0, time.Local)}, want: true},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 8, 1, 0, 0, time.Local)}, want: false},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 19, 0, 0, 0, time.Local)}, want: true},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 20, 0, 0, 0, time.Local)}, want: true},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 21, 0, 0, 0, time.Local)}, want: false},
		{name: "check range", fields: fields{ActiveTime: "* 9-20 * * *"}, args: args{time.Date(2022, 6, 28, 21, 0, 0, 0, time.Local)}, want: false},
		{name: "check list", fields: fields{ActiveTime: "* 9,11 * * *"}, args: args{time.Date(2022, 6, 28, 9, 0, 0, 0, time.Local)}, want: true},
		{name: "check list", fields: fields{ActiveTime: "* 9,11 * * *"}, args: args{time.Date(2022, 6, 28, 10, 0, 0, 0, time.Local)}, want: false},
		{name: "check list", fields: fields{ActiveTime: "* 9,11 * * *"}, args: args{time.Date(2022, 6, 28, 11, 0, 0, 0, time.Local)}, want: true},
		{name: "check invalid", fields: fields{ActiveTime: "* 9,11 * *"}, args: args{time.Date(2022, 6, 28, 11, 0, 0, 0, time.Local)}, want: false},
		{name: "check invalid", fields: fields{ActiveTime: ""}, args: args{time.Date(2022, 6, 28, 11, 0, 0, 0, time.Local)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipe := &Pipeline{
				ActiveTime: tt.fields.ActiveTime,
			}
			if got := pipe.IsActive(tt.args.t); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
