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
			tmpl:    `{"name":{{ref "/username"}}}`,
			srcData: obj([]byte(`{"username":"khs"}`)),
		}, want: obj([]byte(`{"name":"khs"}`))},

		{name: "ref2", args: args{
			tmpl:    `{"name":{{ref "/username"}}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"name":{"given":"hs","family":"k"}}`))},

		{name: "ref3", args: args{
			tmpl:    `{"first":{{ref "/username/given"}},"last":{{ref "/username/family"}}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"first":"hs","last":"k"}`))},

		{name: "refarr", args: args{
			tmpl:    `{"full":{{ref "/username"}}}`,
			srcData: obj([]byte(`{"username":{"given":"hs","family":"k"}}`)),
		}, want: obj([]byte(`{"full":{"given":"hs","family":"k"}}`))},

		{name: "ref-non", args: args{
			tmpl:    `{"first":{{ref "/username/given"}},"last":{{ref "/username/family"}}}`,
			srcData: obj([]byte(`{"username":{"family":"k"}}`)),
		},
			want: obj([]byte(`{"first":null,"last":"k"}`)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.args.tmpl, tt.args.srcData)
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
