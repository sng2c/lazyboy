package tmpl

import (
	"bytes"
	"encoding/json"
	"github.com/lestrrat-go/jspointer"
	"log"
	"text/template"
)

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
