package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/indexer/mocks"
	"github.com/sp0x/torrentd/indexer/status/models"
	"net/http/httptest"
	"testing"
)

func TestServer_Status(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testObj := &Server{}
	reportGen := mocks.NewMockReportGenerator(ctrl)
	//Mocks
	reportGen.EXPECT().GetLatestItems().Return([]models.LatestResult{})
	reportGen.EXPECT().GetIndexesStatus(gomock.Any()).Return([]models.IndexStatus{})
	testObj.status = reportGen
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	testObj.Status(c)

	g.Expect(w.Code).To(gomega.BeEquivalentTo(200))
	var got gin.H
	err := json.Unmarshal(w.Body.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(got).ToNot(gomega.BeNil())
}
