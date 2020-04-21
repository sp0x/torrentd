package indexer

import (
	"bytes"
	"strings"
	"text/template"
)

//Evaluate a template
func applyTemplate(name, tpl string, ctx interface{}) (string, error) {
	funcMap := template.FuncMap{
		"replace": strings.Replace,
	}
	tmpl, err := template.New(name).Funcs(funcMap).Parse(tpl)
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, ctx) //Evaluate the template
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
