package tmpl

import (
	"bytes"
	"encoding/json"
	"github.com/lestrrat-go/jspointer"
	"log"
	"text/template"
)

/*
Data를 json path로 추출해서 쓰려고 함
- json reference 스펙을 템플릿으로 이용하면, json 데이터를 제약없이 치환시킬수 있다.
- 근데 go는 typed 언어라 {"$ref":"path"} 를 primiitve 타입으로 치환하는 것이 매우 어렵다.
- jsref 라이브러리를 이용하려 했는데, 해당 값을 resolve하지 못했을때 error대신 null을 넣게 수정하고 싶었지만 마음이 급해서 잘안되었고,
- reflection 파티가 되는게 좀 그랬다.
- 그리고 interface{} 형태로 들고 처리하는 것이 매우 피곤하다. (매번 캐스팅)
- 그래서 json reference대신 그냥 template에 ref 함수를 추가해서 json path로 치환하게 해서 목적을 달성했다.
*/

func jsonRef(srcData interface{}, path string) string {
	p, err := jspointer.New(path)
	if err != nil {
		return "null"
	}
	got, err := p.Get(srcData)
	if err != nil {
		return "null"
	}
	m, _ := json.Marshal(got)
	return string(m)
}

func Resolve(tmplStr string, srcData interface{}) (string, error) {
	funcMap := template.FuncMap{
		"ref": func(path string) string {
			return jsonRef(srcData, path)
		},
	}

	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, nil)
	if err != nil {
		log.Fatalf("execution: %s", err)
	}
	return tpl.String(), nil
}
