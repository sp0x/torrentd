package indexer

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
)

func TestResolveIndexId_ShouldWorkWithCommaIndexes(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedScope := NewMockScope(ctrl)
	loadedIndexes := make(map[string]IndexCollection)
	aggregateKey := "ix1,ix2,ix3"
	loadedIndexes[aggregateKey] = IndexCollection{}
	mockedScope.EXPECT().Indexes().
		Times(1).
		Return(loadedIndexes)

	rid := ResolveIndexID(mockedScope, "")

	g.Expect(rid).To(gomega.BeEquivalentTo(aggregateKey))
}

func TestResolveIndexId_ShouldWorkWithGlobalIndexes(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedScope := NewMockScope(ctrl)
	loadedIndexes := make(map[string]IndexCollection)
	aggregateKey := "all"
	loadedIndexes[aggregateKey] = IndexCollection{}
	mockedScope.EXPECT().Indexes().
		Times(1).
		Return(loadedIndexes)

	rid := ResolveIndexID(mockedScope, "")

	g.Expect(rid).To(gomega.BeEquivalentTo(aggregateKey))
}
