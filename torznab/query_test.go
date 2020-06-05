package torznab

import (
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/categories"
	"testing"
)

func TestQuery_AddCategory(t *testing.T) {
	g := NewGomegaWithT(t)
	query := Query{}
	g.Expect(query.Categories).To(BeEmpty())
	query.AddCategory(categories.Rental)
	g.Expect(query.Categories).ShouldNot(BeEmpty())
	g.Expect(query.Categories[0]).Should(BeEquivalentTo(categories.Rental.ID))
}
