package indexer

import (
	"testing"
)

func Test_searchBlock_IsSinglePage(t *testing.T) {
	g := gomega.NewWithT(t)
	sb := searchBlock{MaxPages: 1}
	g.Expect(sb.IsSinglePage()).To(gomega.BeTrue())
	sb = searchBlock{MaxPages: 2}
	g.Expect(sb.IsSinglePage()).To(gomega.BeTrue())
	searchInputs := make(map[string]string)
	searchInputs["a"] = "asd"
	sb = searchBlock{MaxPages: 2, Inputs: searchInputs}
	g.Expect(sb.IsSinglePage()).To(gomega.BeFalse())

	sb = searchBlock{MaxPages: 0, Inputs: searchInputs}
	g.Expect(sb.IsSinglePage()).To(gomega.BeFalse())
	//
	sb = searchBlock{MaxPages: 2, Inputs: nil}
	g.Expect(sb.IsSinglePage()).To(gomega.BeTrue())
}
