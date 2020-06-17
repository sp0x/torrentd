package server

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/sp0x/torrentd/indexer"

	"github.com/sp0x/torrentd/config/mocks"
	indexerMocks "github.com/sp0x/torrentd/indexer/mocks"
	httpMocks "github.com/sp0x/torrentd/server/http/mocks"
	"testing"
)

func TestServer_downloadHandler(t *testing.T) {
	//g := NewGomegaWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	config := mocks.NewMockConfig(ctrl)
	context := httpMocks.NewMockContext(ctrl)
	scopeMock := indexerMocks.NewMockScope(ctrl)

	config.EXPECT().GetInt("port").Return(3333).Times(1)
	config.EXPECT().GetString("hostname").Return("").Times(1)
	config.EXPECT().GetBytes("api_key").Return(nil).Times(1)
	s := NewServer(config)
	s.indexerFacade = indexer.NewEmptyFacade(config)
	s.Params.APIKey = []byte("demotoken")

	//Test a simple download request without any params.
	//nothing should happen.
	context.EXPECT().Param("token").Return("")
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().Error(gomock.Any()).Times(0)
	s.downloadHandler(context)

	//Download should error out if the token isn't in JWT
	context.EXPECT().Param("token").Return("demotoken")
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().Error(gomock.Any()).Times(1)
	s.downloadHandler(context)

	//Download should error out if the token doesn't have a valid site.
	tkn := token{}
	tokenString, _ := tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	//Unknown indexer
	context.EXPECT().Error(gomock.Any()).Times(1)
	s.downloadHandler(context)

	//If we're not using any link for the download, just return 404
	tkn = token{Site: "rutracker.org"}
	tokenString, _ = tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	context.EXPECT().String(404, gomock.Any())
	//Unknown indexer
	//context.EXPECT().Error(gomock.Any()).Times(1)
	s.downloadHandler(context)

	//If we've given a valid link, we should see the download
	tkn = token{Site: "rutracker.org", Link: "http://rutracker.org"}
	tokenString, _ = tkn.Encode([]byte("demotoken"))
	context.EXPECT().Param("token").Return(tokenString)
	context.EXPECT().Param("filename").Return("")
	//We expect the url of the site to be checked.
	config.EXPECT().GetSite("rutracker.org")
	config.EXPECT().GetSiteOption("rutracker.org", "url")
	context.EXPECT().Error(errors.New("couldn't open stream for download")).Times(1)
	scopeMock.EXPECT().Lookup(gomock.Any(), "rutracker.org")
	s.indexerFacade.Scope = scopeMock
	s.downloadHandler(context)
}
