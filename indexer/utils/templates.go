package utils

import (
	"bytes"
	"strings"
	"text/template"
)

func GetDefaultFunctionMap() template.FuncMap {
	fmap := template.FuncMap{}
	fmap["replace"] = strings.ReplaceAll
	return fmap
}

// Evaluate a template
func ApplyTemplate(name, templateText string, templateContext interface{}, functions template.FuncMap) (string, error) {
	if functions == nil {
		functions = GetDefaultFunctionMap()
	}
	tmpl, err := template.New(name).Funcs(functions).Parse(templateText)
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, templateContext)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
