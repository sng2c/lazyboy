package tmpl

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

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

func TestResolve(t *testing.T) {

	type args struct {
		tmpl    string
		srcData interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{name: "plain", args: args{
			tmpl:    `{"name":"khs"}`,
			srcData: obj([]byte(`{"username":"khs"}`)),
		}, want: obj([]byte(`{"name":"khs"}`))},

		{name: "ref1", args: args{
			tmpl:    `{"name": #{ $.username } }`,
			srcData: obj([]byte(`{"username":"khs"}`)),
		}, want: obj([]byte(`{"name":"khs"}`))},

		{name: "ref2", args: args{
			tmpl:    `{"name":#{$.username}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"name":{"given":"hs","family":"k"}}`))},

		{name: "ref3", args: args{
			tmpl:    `{"first":#{$.username.given},"last":#{$.username.family}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"first":"hs","last":"k"}`))},

		{name: "refarr", args: args{
			tmpl:    `{"full":#{$.username}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"full":{"given":"hs","family":"k"}}`))},

		{name: "ref-non", args: args{
			tmpl:    `{"first":#{$.username.given},"last":#{$.username.family}}`,
			srcData: obj([]byte(`{"username":{"family":"k"}}`)),
		},
			want: obj([]byte(`{"first":null,"last":"k"}`)),
		},
		{name: "ref-multi", args: args{
			tmpl:    `{"givens":#{$..given}}`,
			srcData: obj([]byte(`[{"username":{"given":"hs","family":"k"}},{"username":{"given":"hanson","family":"k"}}]`)),
		},
			want: obj([]byte(`{"givens":["hs","hanson"]}`)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, _ := NewTemplate(tt.args.tmpl)
			got, err := ResolveTemplate(tmpl, tt.args.srcData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(obj([]byte(got)), tt.want) {
				t.Errorf("Resolve() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransTemplate(t *testing.T) {
	type args struct {
		tmplStr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "trans", args: args{"123"}, want: "123"},
		{name: "trans", args: args{`123 #{ 1 } 456`}, want: `123 {{ref . "1"}} 456`},
		{name: "trans", args: args{`123 #{ 1} 456`}, want: `123 {{ref . "1"}} 456`},
		{name: "trans", args: args{`#{1}`}, want: `{{ref . "1"}}`},
		{name: "trans", args: args{`#{1}`}, want: `{{ref . "1"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transTemplate(tt.args.tmplStr); got != tt.want {
				t.Errorf("transTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
