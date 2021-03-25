package utils

import (
	"bytes"
	"strings"
	"text/template"
)

func GetDefaultFunctionMap() template.FuncMap {
	return template.FuncMap{
		"replace": strings.ReplaceAll,
	}
}

// Evaluate a template
func ApplyTemplate(name, templateText string, ctx interface{}, functions template.FuncMap) (string, error) {
	if functions == nil {
		functions = GetDefaultFunctionMap()
	}
	tmpl, err := template.New(name).Funcs(functions).Parse(templateText)
	if err != nil {
		return "", err
	}
	b := &bytes.Buffer{}
	err = tmpl.Execute(b, ctx) // Evaluate the template
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
