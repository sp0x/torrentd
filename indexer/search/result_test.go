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
	a.ModelData = make(map[string]interface{})
	a.ModelData["a"] = 3
	b = &ScrapeResultItem{}
	b.ModelData = make(map[string]interface{})
	b.ModelData["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

	a = &ScrapeResultItem{}
	a.SourceLink = "asd"
	a.ModelData = make(map[string]interface{})
	a.ModelData["a"] = 3
	b = &ScrapeResultItem{}
	b.SourceLink = "asd"
	b.ModelData = make(map[string]interface{})
	b.ModelData["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

}
