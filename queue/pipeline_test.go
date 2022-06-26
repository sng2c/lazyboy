package queue

import (
	"lazyboy/tmpl"
	"net/http"
	"path"
	"reflect"
	"testing"
	"text/template"
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
				Correction:    1.0,
				queuePath:     path.Join(testBase, "pipeline_test1", "queue1"),
				ReqTmplName:   "_req.json",
				ResTmplName:   "_res.json",
				reqTmplString: "{\n  \"name\": #{name}\n}",
				resTmplString: "{\n  \"body\" : #{$.body}\n}",
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
func newPipeline(configPath string)*Pipeline{
	pipe, err := NewPipelineFromConfigPath(configPath)
	if err != nil {
		return nil
	}
	return pipe
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
		want    *http.Request
		wantErr bool
	}{
		{
			name: "pipe buildReq",
			fields: fields{newQueue(path.Join(testBase, "pipelines", "pipe1", "config.json"))},
			args: , want: , wantErr: },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipe := tt.fields.Pipeline

			got, err := pipe.BuildRequest(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}