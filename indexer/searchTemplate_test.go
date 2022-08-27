package indexer

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/search"
)

func Test_IsComplete_Given_RangeFieldIsUsed_Then_IsComplete_ShouldBeTrue_IfRangeIsComplete(t *testing.T) {
	g := gomega.NewWithT(t)
	data := SearchTemplateData{}
	qry := search.NewQuery()
	var rangeFieldValue search.RangeField = []string{"001", "010"}
	qry.Fields["rangeField"] = rangeFieldValue

	iter := search.NewIterator(qry)

	for i := 1; i < 11; i++ {
		fields, page := iter.Next()
		data.Search = createWorkerJob(nil, nil, fields, page)
		val, _ := data.GetSearchFieldValue("rangeField")
		g.Expect(val).To(gomega.Equal(fmt.Sprintf("%03d", i)))
	}

	g.Expect(iter.IsComplete()).To(gomega.BeTrue())
}

func Test_ApplyField_Given_RangeField_Should_ReturnCorrectValues(t *testing.T) {
	g := gomega.NewWithT(t)
	data := SearchTemplateData{Query: search.NewQuery()}
	rangef := search.RangeField{}
	rangef = append(rangef, []string{"001", "010"}...)
	data.Query.Fields["rangeField"] = rangef
	iter := search.NewIterator(data.Query)

	for i := 1; i < 11; i++ {
		fields, page := iter.Next()
		data.Search = createWorkerJob(nil, nil, fields, page)
		val, _ := data.GetSearchFieldValue("rangeField")
		g.Expect(val).To(gomega.Equal(fmt.Sprintf("%03d", i)))
	}
}
