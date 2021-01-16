package indexer

import (
	"testing"

	"github.com/onsi/gomega"
)

func Test_formatValue_ShouldWorkWithoutFields(t *testing.T) {
	g := gomega.NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "testx"
	value := formatValues(nil, "val{{ .other }}ue", values)
	g.Expect(value).To(gomega.Equal("valtestxue"))
	value = formatValues(nil, "val{{ .x }}ue", values)
	g.Expect(value).To(gomega.Equal("val<no value>ue"))
}

func Test_formatValue_ShouldWorkWithEmptyValues(t *testing.T) {
	g := gomega.NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "test"
	value := formatValues(nil, nil, values)
	g.Expect(value).To(gomega.BeNil())
	value = formatValues(nil, []string{""}, values)
	g.Expect(value).To(gomega.Equal([]string{""}))
}

func Test_formatValue_ShouldWorkWithTextValsWithoutActualValues(t *testing.T) {
	g := gomega.NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "test"
	field := &fieldBlock{}
	field.Block.TextVal = "valx"
	value := formatValues(field, nil, values)
	g.Expect(value).To(gomega.Equal("valx"))
}

func Test_formatValue_ShouldFilterArrays(t *testing.T) {
	g := gomega.NewWithT(t)
	values := make(map[string]interface{})
	values["a"] = "test1"
	values["b"] = "test2"
	field := &fieldBlock{}
	field.Block.TextVal = "valx"
	field.Block.Filters = []filterBlock{
		{
			Name: "number",
			Args: nil,
		},
	}
	value := formatValues(field, []string{
		"val{{.a}}",
		"val{{.b}}",
	}, values)
	valueArray := value.([]string)
	g.Expect(valueArray[0]).To(gomega.Equal("1"))
	g.Expect(valueArray[1]).To(gomega.Equal("2"))
}
