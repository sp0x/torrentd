package indexer

import (
	. "github.com/onsi/gomega"
	"testing"
)

func Test_formatValue_ShouldWorkWithoutFields(t *testing.T) {
	g := NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "test"
	value := formatValues(nil, "val{{ .other }}ue", values)
	g.Expect(value).To(Equal("valtestue"))
	value = formatValues(nil, "val{{ .x }}ue", values)
	g.Expect(value).To(Equal("val<no value>ue"))
}

func Test_formatValue_ShouldWorkWithEmptyValues(t *testing.T) {
	g := NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "test"
	value := formatValues(nil, nil, values)
	g.Expect(value).To(BeNil())
	value = formatValues(nil, []string{""}, values)
	g.Expect(value).To(Equal([]string{""}))
}

func Test_formatValue_ShouldWorkWithTextValsWithoutActualValues(t *testing.T) {
	g := NewWithT(t)
	values := make(map[string]interface{})
	values["other"] = "test"
	field := &fieldBlock{}
	field.Block.TextVal = "valx"
	value := formatValues(field, nil, values)
	g.Expect(value).To(Equal("valx"))
}

func Test_formatValue_ShouldFilterArrays(t *testing.T) {
	g := NewWithT(t)
	values := make(map[string]interface{})
	values["a"] = "test1"
	values["b"] = "test2"
	field := &fieldBlock{}
	field.Block.TextVal = "valx"
	field.Block.Filters = []filterBlock{
		{
			Name: "number",
			Args: nil,
		}}
	value := formatValues(field, []string{
		"val{{.a}}",
		"val{{.b}}",
	}, values)
	valueArray := value.([]string)
	g.Expect(valueArray[0]).To(Equal("1"))
	g.Expect(valueArray[1]).To(Equal("2"))
}
