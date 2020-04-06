package torrent

import (
	"golang.org/x/text/encoding/charmap"
)

func DecodeWindows1251(ba []uint8) []uint8 {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes(ba)
	return out
}
