package torrent

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackpal/bencode-go"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser/encoding"

	"github.com/sp0x/torrentd/indexer"
)

var (
	rxMagnet, _ = regexp.Compile("^(stream-)?magnet:")
	rxHex, _    = regexp.Compile("^[a-f0-9]{40}$")
	rxBase32, _ = regexp.Compile("^[a-z2-7]{32}")
)

func ParseTorrentFromStream(stream io.ReadCloser) (*Definition, error) {
	body, err := ioutil.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	// ioutil.WriteFile("/tmp/rss.torrent", body, os.ModePerm)
	return ParseTorrent(string(body))
}

func ParseTorrentFromURL(h *indexer.Facade, torrentURL string) (*Definition, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", torrentURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := h.Indexer.ProcessRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, errors.New(strconv.Itoa(res.StatusCode))
	}
	return ParseTorrent(string(body))
}

func ParseTorrent(torrent string) (*Definition, error) {
	if rxMagnet.MatchString(torrent) {
		// Torrent is a magnet
		return parseMagnet(torrent)
		//if d.InfoHash == "" {
		//	return nil, errors.New("could not parse magnet torrent id")
		//}
		//return &d, nil
	} else if rxHex.MatchString(torrent) || rxBase32.MatchString(torrent) {
		// if info is a hash (hex/base-32 str)
		return parseMagnet("magnet:?xt=urn:btih:" + torrent)
	} else if len(torrent) == 20 && isTorrentBuff(torrent) {
		return parseMagnet("magnet:?xt=urn:btih:" + torrent)
	} else if isTorrentBuff(torrent) {
		return decodeTorrentBuff([]byte(torrent))
	} else {
		return nil, errors.New("invalid torrent")
	}
}

func isTorrentBuff(buff string) bool {
	return true
}

// Parse a torrent file's content to get more information about it
func decodeTorrentBuff(buff []byte) (*Definition, error) {
	reader := bytes.NewReader(buff)
	var data Definition
	err := bencode.Unmarshal(reader, &data)
	decoder := encoding.GetEncoding("windows1251").NewDecoder()
	if err != nil {
		buff, _ = decoder.Bytes(buff)
		if strings.Contains(string(buff), "<b>премодерация</b>") {
			return nil, errors.New("torrent is still now allowed to be downloaded")
		}
		return nil, err
	}
	buffWriter := &bytes.Buffer{}
	err = bencode.Marshal(buffWriter, data.Info)
	if err != nil {
		log.Warningf("Could not encode torrent info: %v\n", err)
		return nil, err
	}
	data.InfoBuffer = buffWriter.Bytes()
	hash := sha1.New()
	hash.Write(data.InfoBuffer)
	data.InfoHash = fmt.Sprintf("%x", hash.Sum(nil))
	return &data, nil
}

func parseMagnet(m string) (*Definition, error) {
	return nil, nil
}

type RawDefinition struct {
	Announce     string
	AnnounceList [][]string "announce-list" //nolint:govet
	Comment      string
	CreatedBy    string "created by"    //nolint:govet
	CreationDate uint   "creation date" //nolint:govet
	Encoding     string
	Info         DefinitionInfo
	Publisher    string
	PublisherURL string "publisher-url" //nolint:govet
}

type Definition struct {
	Announce     string     "announce"      //nolint:govet
	AnnounceList [][]string "announce-list" //nolint:govet
	Comment      string     "comment"       //nolint:govet
	CreatedBy    string     "created by"    //nolint:govet
	CreationDate uint       "creation date" //nolint:govet
	Encoding     string     "encoding"      //nolint:govet
	Info         DefinitionInfo
	Publisher    string "publisher"     //nolint:govet
	PublisherURL string "publisher-url" //nolint:govet
	InfoBuffer   []byte
	InfoHash     string
}

func (d *Definition) ToMagnetURL() string {
	return fmt.Sprintf("magnet:?xt=urn:btih:%s", d.InfoHash)
}

func (d *Definition) GetTotalFileSize() uint32 {
	files := d.Info.Files
	total := uint32(0)
	for _, f := range files {
		total += uint32(f.Length)
	}
	return total
}

type DefinitionInfo struct {
	FileDuration []int               "file-duration" //nolint:govet
	FileMedia    []int               "file-media"    //nolint:govet
	Files        []DefinitionFile    "files"         //nolint:govet
	Name         string              "name"          //nolint:govet
	PieceLength  uint                "piece length"  //nolint:govet
	Pieces       string              "pieces"        //nolint:govet
	Profiles     []DefinitionProfile "profiles"      //nolint:govet
}

type DefinitionFile struct {
	Length uint64   "length" //nolint:govet
	Path   []string "path"   //nolint:govet
}

type DefinitionProfile struct {
	ACodec string "acodec" //nolint:govet
	Height uint   "height" //nolint:govet
	VCodec string "vcodec" //nolint:govet
	Width  uint   "width"  //nolint:govet
}
