package queue

import (
	"lazyboy/tmpl"
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
		want    *Queue
		wantErr bool
	}{
		{name: "queue_test1", args: args{configPath: path.Join(testBase, "queue_test1", "queue1", "_config.json")},
			want: &Queue{
				TakePerTick:     3,
				Correction:      1.0,
				QueuePath:       path.Join(testBase, "queue_test1", "queue1"),
				ReqTmplFilename: "_req.json",
				ResTmplFilename: "_res.json",
				ReqTmplString:   "{\n  \"name\": #{name}\n}",
				ResTmplString:   "{\n  \"body\" : #{$.body}\n}",
			}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewQueueFromConfigPath(tt.args.configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewQueueFromConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueueFromConfigPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func newQueue(configPath string) *Queue {
	queue, err := NewQueueFromConfigPath(configPath)
	if err != nil {
		return nil
	}
	return queue
}

func TestQueue_Take(t *testing.T) {
	type fields struct {
		Queue *Queue
	}
	type args struct {
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		{name: "take_test1", fields: fields{newQueue(path.Join(testBase, "queue_take_test1", "queue1", "_config.json"))},
			args: args{},
			want: [][]byte{
				[]byte("0"),
				[]byte("1"),
				[]byte("2"),
			}},
		{name: "take_test1", fields: fields{newQueue(path.Join(testBase, "queue_take_test1", "queue1", "_config.json"))},
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
				t.Errorf("Queue failed construction")
			}
			if got := q.Take(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Take() = \n%v, \nwant \n%v", got, tt.want)
			}
		})
	}
}
