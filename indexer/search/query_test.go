package search

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer/categories"
)

func TestQuery_AddCategory(t *testing.T) {
	g := NewGomegaWithT(t)
	query := Query{}
	g.Expect(query.Categories).To(BeEmpty())

	query.AddCategory(categories.Rental)

	g.Expect(query.Categories).ShouldNot(BeEmpty())
	g.Expect(query.Categories[0]).Should(BeEquivalentTo(categories.Rental.ID))
}

func TestParseQueryString_GivenSimpleStringThenQFieldShouldBeUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	q := NewSearchFromQuery("simple query")

	g.Expect(q.QueryString).ToNot(BeEmpty())
	g.Expect(q.QueryString).To(Equal("simple query"))
}

func TestParseQueryString_Given_RangeQuery_Then_ItShouldBeParsed(t *testing.T) {
	g := NewGomegaWithT(t)

	q := NewSearchFromQuery("$phone:range(1, 200)")

	g.Expect(q.QueryString).To(BeEmpty())
	g.Expect(q.Fields).ToNot(BeNil())
	g.Expect(q.Fields["phone"]).ToNot(BeNil())
	g.Expect(q.Fields["phone"]).To(BeAssignableToTypeOf(RangeField{}))
	rangeField := q.Fields["phone"].(RangeField)
	g.Expect(rangeField[0]).To(BeEquivalentTo("1"))
	g.Expect(rangeField[1]).To(BeEquivalentTo("200"))
}
