package indexer

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/search"
)

func Test_RangeValue_ShouldChangeAfterEveryCall(t *testing.T) {
	g := gomega.NewWithT(t)
	data := SearchTemplateData{}
	qry := search.NewQuery()
	rangef := search.RangeField{}
	rangef = append(rangef, []string{"001", "010"}...)
	qry.Fields["rangeField"] = rangef

	srch := search.NewSearch(qry)
	data.Context = RunContext{
		Search: srch.(*search.Search),
	}

	for i := 1; i < 11; i++ {
		val := data.RangeValue("rangeField")
		g.Expect(val).To(gomega.Equal(fmt.Sprintf("%03d", i)))
	}
}

func Test_ApplyField_Given_RangeField_Should_ReturnCorrectValues(t *testing.T) {
	g := gomega.NewWithT(t)
	data := SearchTemplateData{Query: search.NewQuery()}
	rangef := search.RangeField{}
	rangef = append(rangef, []string{"001", "010"}...)
	data.Query.Fields["rangeField"] = rangef
	srch := search.NewSearch(data.Query)
	data.Context = RunContext{
		Search: srch.(*search.Search),
	}

	for i := 1; i < 11; i++ {
		val, _ := data.ApplyField("rangeField")
		g.Expect(val).To(gomega.Equal(fmt.Sprintf("%03d", i)))
	}
}
