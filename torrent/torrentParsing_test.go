package torrent

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/indexer"
)

func Test_ParseTorrentFromURL(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	index := indexer.NewMockIndexer(ctrl)
	torrentURL := "http://torrentLink.com/torrent"
	responseProxy, pipeWr := indexer.NewResponseProxy()
	torrentBody := getTorrentBuffer()
	go func() {
		responseProxy.ContentLengthChan <- int64(len(torrentBody))
		_, _ = pipeWr.Write(torrentBody)
		_ = pipeWr.Close()
	}()
	index.EXPECT().Download(torrentURL).Return(responseProxy, nil)
	def, err := ParseTorrentFromURL(index, torrentURL)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(def).ToNot(gomega.BeNil())
	magnetURL := def.ToMagnetURL()
	torrentSize := def.GetTotalFileSize()
	g.Expect(def.Announce).To(gomega.Equal("http://bttracker.debian.org:6969/announce"))
	g.Expect(def.Comment).To(gomega.Equal("\"Debian CD from cdimage.debian.org\""))
	g.Expect(def.Info.Name).To(gomega.Equal("debian-10.9.0-amd64-netinst.iso"))
	g.Expect(def.Info.PieceLength).To(gomega.Equal(uint(262144)))
	g.Expect(magnetURL).ToNot(gomega.BeNil())
	g.Expect(magnetURL).To(gomega.Equal("magnet:?xt=urn:btih:3b1cd97eb93cf0921b8abbf9a84cd811eccbb691"))
	g.Expect(torrentSize).To(gomega.Equal(uint32(0)))
	g.Expect(def.Info.Files).To(gomega.BeNil())
}

func getTorrentBuffer() []byte {
	buf := bytes.NewBuffer(nil)
	testFile, err := os.Open(path.Join("testdata", "sample.torrent"))
	if err != nil {
		return nil
	}
	defer func() {
		_ = testFile.Close()
	}()
	_, _ = io.Copy(buf, testFile)
	return buf.Bytes()
}
