package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"regexp"
)

var rxMagnet, _ = regexp.Compile("^(stream-)?magnet:")
var rxHex, _ = regexp.Compile("^[a-f0-9]{40}$")
var rxBase32, _ = regexp.Compile("^[a-z2-7]{32}")

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
	if err != nil {
		return nil, err
	}
	dataBuff := []byte{}
	buffWriter := bytes.NewBuffer(dataBuff)
	err = bencode.Marshal(buffWriter, &data.Info)
	if err != nil {
		return nil, err
	}
	data.InfoBuffer = dataBuff
	hash := sha1.New()
	hash.Write(dataBuff)
	data.InfoHash = fmt.Sprintf("%x", hash.Sum(nil))
	return &data, nil
}

func parseMagnet(m string) (*Definition, error) {
	return nil, nil
}

type Definition struct {
	Announce     string
	AnnounceList [][]string "announce-list"
	Comment      string
	CreatedBy    string "created by"
	CreationDate uint   "creation date"
	Encoding     string
	Info         DefinitionInfo
	Publisher    string
	PublisherUrl string "publisher-url"
	InfoBuffer   []byte
	InfoHash     string
}

type DefinitionInfo struct {
	FileDuration []int "file-duration"
	FileMedia    []int "file-media"
	Files        []DefinitionFile
	Name         string
	PieceLength  uint "piece length"
	Pieces       string
	Profiles     []DefinitionProfile
}

type DefinitionFile struct {
	Length uint64
	Path   []string
}

type DefinitionProfile struct {
	ACodec string "acodec"
	Height uint   "height"
	VCodec string "vcodec"
	Width  uint   "width"
}
