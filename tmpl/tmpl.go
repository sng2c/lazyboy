package tmpl

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"text/template"
)

/*
Data를 json path로 추출해서 쓰려고 함
- json reference 스펙을 템플릿으로 이용하면, json 데이터를 제약없이 치환시킬수 있다.
- json을 json을 위한 표준만으로 pointing하고 replace하는 것이 아주 편리하겠다는 생각을 함.

- 근데 go는 typed 언어라 {"$ref":"path"} 를 primiitve 타입으로 치환하는 것이 매우 어렵다.
- jsref 라이브러리를 이용하려 했는데, 해당 값을 resolve하지 못했을때 error대신 null을 넣게 수정하고 싶었지만 마음이 급해서 잘안되었고,
- reflection 파티가 되는게 좀 그랬다.
- 그리고 json 데이터의 항목들이 유동적일거라 interface{} 형태로 들고 처리해야하는데 이게 매우 피곤하다. (매번 타입확인 후 캐스팅)

- JSON의 자료형은 JS의 기준이고 이걸 go에 끼워맞추니까 marshal 타이밍에 타입이 결정되어 바꾸기 어렵게 된다.
- 그래서 그냥 JSON자료형의 구조체인 상태로 다루면 수월한데, jsonvalue 라이브러리가 적당했다. https://github.com/Andrew-M-C/go.jsonvalue
- 근데 이걸 선택하면, json path를 찾는것부터 치환하는것까지 내가 직접 구현해야 해서 크게 이득이 없다.

- 그래서 json reference대신 그냥 template에 jspointer로 개별치환하는 ref 함수를 추가해서 json path로 치환하게 해서 목적을 달성했다.
- jspointer 대신 jsonpath를 써야겠다. -> 완료
- jsonpath 문법은 https://goessner.net/articles/JsonPath/ 참고

*/

var builder = gval.Full(jsonpath.PlaceholderExtension())

func jsonRefJs(pathStr string, tmplData interface{}) string {
	ref := jsonRef(pathStr, tmplData)
	if ref == nil {
		return "null"
	}
	m, _ := json.Marshal(ref)
	return string(m)
}

func jsonRef(pathStr string, tmplData interface{}) interface{} {
	path, err := builder.NewEvaluable(pathStr)
	if err != nil {
		log.Println(err)
		return nil
	}
	got, err := path(context.Background(), tmplData)
	if err != nil {
		log.Println(err)
		return nil
	}
	return got
	//return got.(string)
}

func jsonRefText(pathStr string, tmplData interface{}) string {
	ref := jsonRef(pathStr, tmplData)
	if ref == nil {
		return ""
	}
	return ref.(string)
}

func jsonRefQuote(pathStr string, tmplData interface{}) string {
	ref := jsonRef(pathStr, tmplData)
	if ref == nil {
		return ""
	}
	return strconv.Quote(ref.(string))
}

func trim(q string, s string) string {
	return strings.Trim(s, q)
}

func NewTemplate(tmplStr string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"ref":      jsonRef,
		"refjs":    jsonRefJs,
		"reftext":  jsonRefText,
		"refquote": jsonRefQuote,
		"trim":     trim,
	}
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		log.Printf("parsing: %s", err)
		return nil, err
	}
	return tmpl, nil
}

func ResolveTemplate(tmpl *template.Template, srcData interface{}) ([]byte, error) {
	var outbuf bytes.Buffer
	err := tmpl.Execute(&outbuf, srcData)
	if err != nil {
		log.Printf("execution: %s", err)
		return nil, err
	}
	return outbuf.Bytes(), nil
}
