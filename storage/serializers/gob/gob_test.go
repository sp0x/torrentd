package gob

import (
	"github.com/sp0x/torrentd/storage/internal"
	"testing"
)

func TestGob(t *testing.T) {
	internal.SerializerTester(t, Serializer)
}
