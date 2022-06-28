package proc

import (
	"bytes"
	"lazyboy/queue"
	"testing"
)

func TestBuildHttpRequest(t *testing.T) {
	type args struct {
		req *queue.Req
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "ReqBuild", args: args{req: &queue.Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   nil,
			BodyType:  queue.BodyTypeText,
			BodyStr:   "HELLO",
			BodyObj:   nil,
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 5\r\n\r\nHELLO", wantErr: false},
		{name: "ReqBuild", args: args{req: &queue.Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   [][]string{{"X-TEST", "WORLD"}},
			BodyType:  queue.BodyTypeText,
			BodyStr:   "HELLO",
			BodyObj:   nil,
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 5\r\nX-Test: WORLD\r\n\r\nHELLO", wantErr: false},
		{name: "ReqBuild", args: args{req: &queue.Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   [][]string{{"X-TEST", "WORLD"}},
			BodyType:  queue.BodyTypeJson,
			BodyStr:   "",
			BodyObj:   map[string]string{"last_name": "hs", "first_name": "k"},
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 35\r\nX-Test: WORLD\r\n\r\n{\"first_name\":\"k\",\"last_name\":\"hs\"}", wantErr: false},
		{name: "ReqBuild", args: args{req: &queue.Req{
			Method:    "GET",
			Url:       "http://localhost?abc=123",
			Headers:   [][]string{{"X-TEST", "WORLD"}},
			BodyType:  queue.BodyTypeJson,
			BodyStr:   "",
			BodyObj:   nil,
			BodyBytes: nil,
		}}, want: "GET /?abc=123 HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nX-Test: WORLD\r\n\r\nnull", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildHttpRequest(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildHttpRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			buf := bytes.NewBuffer([]byte{})
			_ = got.Write(buf)
			reqStr := buf.String()
			if reqStr != tt.want {
				t.Errorf("BuildHttpRequest() got = %#v, want %#v", reqStr, tt.want)
			}
		})
	}
}
