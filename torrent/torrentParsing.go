package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser/encoding"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var rxMagnet, _ = regexp.Compile("^(stream-)?magnet:")
var rxHex, _ = regexp.Compile("^[a-f0-9]{40}$")
var rxBase32, _ = regexp.Compile("^[a-z2-7]{32}")

func ParseTorrentFromStream(stream io.ReadCloser) (*Definition, error) {
	body, err := ioutil.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	//ioutil.WriteFile("/tmp/rss.torrent", body, os.ModePerm)
	return ParseTorrent(string(body))
}

func ParseTorrentFromUrl(h *TorrentHelper, torrentUrl string) (*Definition, error) {
	req, _ := http.NewRequest("GET", torrentUrl, nil)
	res, err := h.indexer.ProcessRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return nil, errors.New(strconv.Itoa(res.StatusCode))
	}
	return ParseTorrent(string(body))
}

func ParseTorrent(torrent string) (*Definition, error) {
	if rxMagnet.MatchString(torrent) {
		//Torrent is a magnet
		return parseMagnet(torrent)
		//if d.InfoHash == "" {
		//	return nil, errors.New("could not parse magnet torrent id")
		//}
		//return &d, nil
	} else if rxHex.MatchString(torrent) || rxBase32.MatchString(torrent) {
		// if info is a hash (hex/base-32 str)
		return parseMagnet("magnet:?xt=urn:btih:" + torrent)
	} else if len(torrent) == 20 && isTorrentBuff(torrent) {
		//if .torrent file buffer
		return parseMagnet("magnet:?xt=urn:btih:" + torrent)
	} else if isTorrentBuff(torrent) {
		return decodeTorrentBuff([]byte(torrent))
	} else {
		return nil, errors.New("invalid torrent")
	}

	return nil, nil
}

func isTorrentBuff(buff string) bool {
	return true
}

//Parse a torrent file's content to get more information about it
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
	AnnounceList [][]string "announce-list"
	Comment      string
	CreatedBy    string "created by"
	CreationDate uint   "creation date"
	Encoding     string
	Info         DefinitionInfo
	Publisher    string
	PublisherUrl string "publisher-url"
}

type Definition struct {
	Announce     string     "announce"
	AnnounceList [][]string "announce-list"
	Comment      string     "comment"
	CreatedBy    string     "created by"
	CreationDate uint       "creation date"
	Encoding     string     "encoding"
	Info         DefinitionInfo
	Publisher    string "publisher"
	PublisherUrl string "publisher-url"
	InfoBuffer   []byte
	InfoHash     string
}

func (d *Definition) ToMagnetUrl() string {
	return fmt.Sprintf("magnet:?xt=urn:btih:%s", d.InfoHash)
}

func (d *Definition) GetTotalFileSize() uint64 {
	files := d.Info.Files
	total := uint64(0)
	for _, f := range files {
		total += f.Length
	}
	return total
}

type DefinitionInfo struct {
	FileDuration []int               "file-duration"
	FileMedia    []int               "file-media"
	Files        []DefinitionFile    "files"
	Name         string              "name"
	PieceLength  uint                "piece length"
	Pieces       string              "pieces"
	Profiles     []DefinitionProfile "profiles"
}

type DefinitionFile struct {
	Length uint64   "length"
	Path   []string "path"
}

type DefinitionProfile struct {
	ACodec string "acodec"
	Height uint   "height"
	VCodec string "vcodec"
	Width  uint   "width"
}
