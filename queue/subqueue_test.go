package queue

import (
	"log"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	//setup()
	log.Println(">>>>>>>>>>>>>>>>>")
	code := m.Run()
	//shutdown()
	log.Println("<<<<<<<<<<<<<<<<<")
	os.Exit(code)
}
func Test_chooseValidDataFilepath(t *testing.T) {

	type args struct {
		queuePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "choose_test1", args: args{"../test/subqueue_offer_test1"}, want: "../test/subqueue_offer_test1/1.json", wantErr: false},
		{name: "choose_test2", args: args{"../test/subqueue_offer_test2"}, want: "../test/subqueue_offer_test2/1.json", wantErr: false},
		{name: "choose_test3", args: args{"../test/subqueue_offer_test3"}, want: "", wantErr: true},
		{name: "choose_test4", args: args{"../test/subqueue_offer_test4"}, want: "", wantErr: true},
		{name: "choose_test5", args: args{"../test/subqueue_offer_test5"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OfferSubQueue(tt.args.queuePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OfferSubQueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				gotPath := path.Join(got.QueuePath, got.SubQueueName)
				if gotPath != tt.want {
					t.Errorf("OfferSubQueue() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSubQueue_Take(t *testing.T) {
	// clone test dir into sandbox

	type fields struct {
		QueuePath    string
		SubQueueName string
		Pos          SubQueuePos
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subq := &SubQueue{
				QueuePath:    tt.fields.QueuePath,
				SubQueueName: tt.fields.SubQueueName,
				Pos:          tt.fields.Pos,
			}
			if got := subq.Take(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Take() = %v, want %v", got, tt.want)
			}
		})
	}
}
