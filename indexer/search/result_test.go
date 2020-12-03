package search

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestExternalResultItem_Equals(t *testing.T) {
	g := gomega.NewWithT(t)
	a := &ScrapeResultItem{}
	b := &ScrapeResultItem{}
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

	a = &ScrapeResultItem{}
	a.ExtraFields = make(map[string]interface{})
	a.ExtraFields["a"] = 3
	b = &ScrapeResultItem{}
	b.ExtraFields = make(map[string]interface{})
	b.ExtraFields["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

	a = &ScrapeResultItem{}
	a.Title = "asd"
	a.ExtraFields = make(map[string]interface{})
	a.ExtraFields["a"] = 3
	b = &ScrapeResultItem{}
	b.Title = "asd"
	b.ExtraFields = make(map[string]interface{})
	b.ExtraFields["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

}
