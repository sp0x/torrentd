package server

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sp0x/torrentd/config/mocks"
	"github.com/sp0x/torrentd/indexer"
	httpMocks "github.com/sp0x/torrentd/server/http/mocks"
)

func prepareTestServer(ctrl *gomock.Controller, config *mocks.MockConfig) (*Server, *httpMocks.MockContext) {
	context := httpMocks.NewMockContext(ctrl)

	config.EXPECT().GetInt("port").Return(3333).Times(1)
	config.EXPECT().GetString("hostname").Return("").Times(1)
	config.EXPECT().GetBytes("api_key").Return(nil).Times(1)

	server := NewServer(config)
	server.indexerFacade = indexer.NewEmptyFacade(config)
	server.Params.APIKey = []byte("demotoken")

	return server, context
}

func TestServer_downloadHandler_Should_Work_WithoutAnyParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	server, context := prepareTestServer(ctrl, config)

	// Test a simple download request without any params.
	// nothing should happen.
	context.EXPECT().Param("token").Return("")
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().Error(gomock.Any()).Times(0)
	server.downloadHandler(context)
}

func TestServer_downloadHelper_Should_ErrorOut_If_TokenIsNotJWT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	server, context := prepareTestServer(ctrl, config)

	// Download should error out if the token isn't in JWT
	context.EXPECT().Param("token").Return("demotoken")
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().Error(gomock.Any()).Times(1)
	server.downloadHandler(context)
}

func TestServer_downloadHelper_Should_ErrorOut_If_TokenIsInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	server, context := prepareTestServer(ctrl, config)

	// Download should error out if the token doesn't have a valid site.
	tkn := token{}
	tokenString, _ := tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	// Unknown indexer
	context.EXPECT().String(404, gomock.Any())
	server.downloadHandler(context)
}

func TestServer_downloadHandler_Should_ErrorOut_If_FilenameValueIsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	server, context := prepareTestServer(ctrl, config)

	// If we're not using any link for the download, just return 404
	tkn := token{IndexName: "rutracker.org"}
	tokenString, _ := tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().String(404, gomock.Any())
	// Unknown indexer
	// context.EXPECT().Error(gomock.Any()).Times(1)
	server.downloadHandler(context)
}

func TestServer_downloadHandler(t *testing.T) {
	// g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	server, context := prepareTestServer(ctrl, config)

	// If we've given a valid link, we should see the download
	scopeMock := indexer.NewMockScope(ctrl)
	mockedIndexer := indexer.NewMockIndexer(ctrl)
	tkn := token{IndexName: "rutracker.org", Link: "http://rutracker.org"}
	responseProxy, pipeWr := indexer.NewResponseProxy()
	go func() {
		responseProxy.ContentLengthChan <- 6
		_, _ = pipeWr.Write([]byte("result"))
		_ = pipeWr.Close()
	}()

	tokenString, _ := tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	// We expect the url of the site to be checked.
	// context.EXPECT().Error(errors.New("couldn't open stream for download")).Times(1)
	context.EXPECT().Header("Content-Type", gomock.Any())
	context.EXPECT().Header("Content-Disposition", gomock.Any())
	context.EXPECT().Header("Content-Transfer-Encoding", gomock.Any())
	context.EXPECT().DataFromReader(200, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	scopeMock.EXPECT().Lookup(gomock.Any(), "rutracker.org").Return(mockedIndexer, nil)
	mockedIndexer.EXPECT().Download(tkn.Link).
		Return(responseProxy, nil)

	server.indexerFacade.IndexScope = scopeMock
	server.downloadHandler(context)
}
