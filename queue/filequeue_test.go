package queue

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_OfferFileQueue(t *testing.T) {

	type args struct {
		queuePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "choose_test1", args: args{path.Join(testBase, "filequeue_offer_test1")}, want: path.Join(testBase, "filequeue_offer_test1", "1.jsonl"), wantErr: false},
		{name: "choose_test2", args: args{path.Join(testBase, "filequeue_offer_test2")}, want: path.Join(testBase, "filequeue_offer_test2", "1.jsonl"), wantErr: false},
		{name: "choose_test3", args: args{path.Join(testBase, "filequeue_offer_test3")}, want: "", wantErr: true},
		{name: "choose_test4", args: args{path.Join(testBase, "filequeue_offer_test4")}, want: "", wantErr: true},
		{name: "choose_test5", args: args{path.Join(testBase, "filequeue_offer_test5")}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OfferFileQueue(tt.args.queuePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OfferFileQueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				gotPath := path.Join(got.QueuePath, got.FileQueueName)
				if gotPath != tt.want {
					t.Errorf("OfferFileQueue() got = %v, want %v", gotPath, tt.want)
				}
			}
		})
	}
}

func TestFileQueue_Take(t *testing.T) {
	// clone test dir into sandbox

	type fields struct {
		QueuePath    string
		SubQueueName string
	}
	type args struct {
		n int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		// TODO: Add test cases.
		{
			name: "data1",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data1.jsonl",
			},
			args: args{2},
			want: [][]byte{[]byte(`{"seq": 0}`), []byte(`{"seq": 1}`)},
		},
		{
			name: "data1",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data1.jsonl",
			},
			args: args{1},
			want: [][]byte{[]byte(`{"seq": 2}`)},
		},
		{
			name: "data1",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data1.jsonl",
			},
			args: args{10},
			want: [][]byte{[]byte(`{"seq": 3}`), []byte(`{"seq": 4}`)},
		},
		{
			name: "data1",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data1.jsonl",
			},
			args: args{1},
			want: [][]byte{},
		},
		{
			name: "data2",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data2.jsonl",
			},
			args: args{3},
			want: [][]byte{
				[]byte("0"),
				[]byte("1"),
				[]byte("2"),
			},
		},
		{
			name: "data2",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data2.jsonl",
			},
			args: args{4},
			want: [][]byte{},
		},
		{
			name: "data3",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data3.jsonl",
			},
			args: args{5},
			want: [][]byte{
				[]byte("0"),
				[]byte("1"),
				[]byte("2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subq, _ := NewFileQueue(tt.fields.QueuePath, tt.fields.SubQueueName)
			if got := subq.Take(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Take() = %v, want %v", got, tt.want)
			}
		})
	}

	filename := path.Join(testBase, "filequeue_take_test1", "data2.jsonl")
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0)
	if err != nil {
		log.Println(err)
	}
	file.Write([]byte("3\n"))
	file.Close()
	cont, _ := os.ReadFile(filename)
	log.Println(cont)

	tests2 := []struct {
		name   string
		fields fields
		args   args
		want   [][]byte
	}{
		{
			name: "data2",
			fields: fields{
				QueuePath:    path.Join(testBase, "filequeue_take_test1"),
				SubQueueName: "data2.jsonl",
			},
			args: args{4},
			want: [][]byte{[]byte("3")},
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			subq, _ := NewFileQueue(tt.fields.QueuePath, tt.fields.SubQueueName)
			if got := subq.Take(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Take() = %v, want %v", got, tt.want)
			}
		})
	}
}
