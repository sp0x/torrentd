package torrent

import (
	"bytes"
	"errors"
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

//Parse a torrent file content
func decodeTorrentBuff(buff []byte) (*Definition, error) {
	reader := bytes.NewReader(buff)
	var data interface{}
	err := bencode.Unmarshal(reader, &data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func parseMagnet(m string) (*Definition, error) {
	return nil, nil
}

type Definition struct {
	InfoHash string
}
