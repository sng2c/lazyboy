package queue

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func newHeader(hdr ...[]string) http.Header {
	header := http.Header{}
	for _, h := range hdr {
		header.Add(h[0], h[1])
	}
	return header
}

func TestBuildHttpRequest(t *testing.T) {
	type args struct {
		req *Req
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "ReqBuild", args: args{req: &Req{
			Method:  "GET",
			Url:     "http://localhost",
			Headers: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\n\r\n", wantErr: false},
		{name: "ReqBuild", args: args{req: &Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   nil,
			BodyType:  BodyTypeText,
			BodyStr:   "HELLO",
			BodyJson:  nil,
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 5\r\n\r\nHELLO", wantErr: false},
		{name: "ReqBuild", args: args{req: &Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   map[string]interface{}{"X-TEST": "WORLD"},
			BodyType:  BodyTypeText,
			BodyStr:   "HELLO",
			BodyJson:  nil,
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 5\r\nX-Test: WORLD\r\n\r\nHELLO", wantErr: false},
		{name: "ReqBuild", args: args{req: &Req{
			Method:    "GET",
			Url:       "http://localhost",
			Headers:   map[string]interface{}{"X-TEST": "WORLD"},
			BodyType:  BodyTypeJson,
			BodyStr:   "",
			BodyJson:  map[string]string{"last_name": "hs", "first_name": "k"},
			BodyBytes: nil,
		}}, want: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 35\r\nX-Test: WORLD\r\n\r\n{\"first_name\":\"k\",\"last_name\":\"hs\"}", wantErr: false},
		{name: "ReqBuild", args: args{req: &Req{
			Method:    "GET",
			Url:       "http://localhost?abc=123",
			Headers:   map[string]interface{}{"X-TEST": "WORLD"},
			BodyType:  BodyTypeJson,
			BodyStr:   "",
			BodyJson:  nil,
			BodyBytes: nil,
		}}, want: "GET /?abc=123 HTTP/1.1\r\nHost: localhost\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nX-Test: WORLD\r\n\r\nnull", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.req.BuildHttpRequest()
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

func Test_BuildResTemplate(t *testing.T) {
	type args struct {
		resTmpl  *template.Template
		response *http.Response
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "parse response", args: args{
			resTmpl: _convTemplate(`{"httpCode":{{refjs "$.StatusCode" . }}}`),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "http",
			},
		}, want: []byte(`{"httpCode":200}`), wantErr: false},
		{name: "parse response", args: args{
			resTmpl: _convTemplate(`{"httpCode":{{refjs "$.StatusCode" . }}}`),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "http",
			},
		}, want: []byte(`{"httpCode":200}`), wantErr: false},
		{name: "parse response", args: args{
			resTmpl: _convTemplate(`{"userName":{{refjs "$.BodyJson.name" .}}}`),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "http",
				Header:     newHeader([]string{"Content-type", "application/json"}),
				Body:       io.NopCloser(strings.NewReader(`{"name":"sng2c"}`)),
			},
		}, want: []byte(`{"userName":"sng2c"}`), wantErr: false},
		{name: "parse response", args: args{
			resTmpl: _convTemplate(`{"greetings":"hello {{reftext "$.BodyText" .}}!"}`),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "http",
				Header:     newHeader([]string{"Content-type", "text/html"}),
				Body:       io.NopCloser(strings.NewReader(`sng2c`)),
			},
		}, want: []byte(`{"greetings":"hello sng2c!"}`), wantErr: false},
		{name: "parse response", args: args{
			resTmpl: _convTemplate(`{"greetings":"hello {{reftext "$.BodyText" .}}!"}`),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "http",
				Header:     newHeader([]string{"Content-type", "text/html"}),
				Body:       io.NopCloser(strings.NewReader("sng2c\n?")),
			},
		}, want: []byte("{\"greetings\":\"hello sng2c\n?!\"}"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := NewResFromHttpResponse(tt.args.response, BodyTypeNone)
			got, err := BuildResTemplate(tt.args.resTmpl, res)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildResTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildResTemplate() got = %#v, want %#v", string(got), string(tt.want))
			}
		})
	}
}
