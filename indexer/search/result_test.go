package search

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestExternalResultItem_Equals(t *testing.T) {
	g := gomega.NewWithT(t)
	a := &ExternalResultItem{}
	b := &ExternalResultItem{}
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

	a = &ExternalResultItem{}
	a.ExtraFields = make(map[string]interface{})
	a.ExtraFields["a"] = 3
	b = &ExternalResultItem{}
	b.ExtraFields = make(map[string]interface{})
	b.ExtraFields["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

	a = &ExternalResultItem{}
	a.Title = "asd"
	a.ExtraFields = make(map[string]interface{})
	a.ExtraFields["a"] = 3
	b = &ExternalResultItem{}
	b.Title = "asd"
	b.ExtraFields = make(map[string]interface{})
	b.ExtraFields["a"] = 3
	g.Expect(a.Equals(b)).To(gomega.BeTrue())

}
