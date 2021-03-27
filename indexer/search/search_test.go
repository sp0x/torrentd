package search

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestNewSearch(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	rangef := make(RangeField, 2)
	rangef[0] = "0"
	rangef[1] = "2"
	q.Fields["field"] = rangef

	s := NewSearch(q)
	g.Expect(s.HasFieldState()).To(gomega.BeTrue())
	rangeState, err := s.GetFieldState("field", nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(rangeState).ToNot(gomega.BeNil())

	g.Expect(s.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(s.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(s.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(s.HasNext()).To(gomega.BeFalse())
}
