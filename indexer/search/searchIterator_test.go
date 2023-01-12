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

	state := NewIterator(q)
	// g.Expect(s.HasFieldState()).To(gomega.BeTrue())
	rangeState, err := state.GetFieldStateOrDefault("field", nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(rangeState).ToNot(gomega.BeNil())

	g.Expect(rangeState.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(rangeState.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(rangeState.HasNext()).To(gomega.BeTrue())
	rangeState.Next()
	g.Expect(rangeState.HasNext()).To(gomega.BeFalse())
}

func Test_Given_QueryPageCountEqualsPage_Then_IsComplete_Should_BeTrue(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 1
	q.Page = 10

	s := NewIterator(q)
	s.CurrentPage = 11

	complete := s.IsComplete()
	g.Expect(complete).To(gomega.Equal(true))
}

func Test_Given_QueryPagesNotExceeded_Then_IsComplete_Should_BeFlase(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 10
	q.Page = 5

	s := NewIterator(q)

	complete := s.IsComplete()
	g.Expect(complete).To(gomega.Equal(false))
}

func Test_Given_QueryPagesAreUnlimited_Then_IsComplete_Should_BeFalse(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 0
	q.Page = 103

	s := NewIterator(q)

	complete := s.IsComplete()
	g.Expect(complete).To(gomega.Equal(false))
}

func Test_Given_StaleResultsAndQueryConfiguredSo_Then_IsComplete_Should_BeTrue(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.StopOnStale = true
	q.NumberOfPagesToFetch = 100
	q.Page = 1

	iter := NewIterator(q)
	results := []ResultItemBase{
		&ScrapeResultItem{
			isNew: true,
		},
		&ScrapeResultItem{
			isNew: false,
		},
	}
	iter.UpdateIteratorState(results)

	complete := iter.IsComplete()
	g.Expect(complete).To(gomega.Equal(true))
}

func TestNewIterator(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 100
	q.Page = 1

	iter := NewIterator(q)

	g.Expect(iter).ToNot(gomega.BeNil())
}

func TestNewIterator_Next_IteratesPages(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 100
	q.Page = 1

	iter := NewIterator(q)
	_, page := iter.Next()

	g.Expect(iter).ToNot(gomega.BeNil())
	g.Expect(page).To(gomega.Equal(uint(1)))
}

func TestNewIterator_Next_IteratesPagesWithCompletion(t *testing.T) {
	g := gomega.NewWithT(t)
	q := NewQuery()
	q.NumberOfPagesToFetch = 2
	q.Page = 1

	iter := NewIterator(q)
	_, page := iter.Next()
	_, page2 := iter.Next()

	g.Expect(iter).ToNot(gomega.BeNil())
	g.Expect(page).To(gomega.Equal(uint(1)))
	g.Expect(page2).To(gomega.Equal(uint(2)))
	g.Expect(iter.IsComplete()).To(gomega.BeTrue())
}
