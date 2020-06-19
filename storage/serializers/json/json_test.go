package json

import (
	"github.com/sp0x/torrentd/storage/internal"
	"testing"
)

func TestJSON(t *testing.T) {
	internal.SerializerTester(t, Serializer)
}
