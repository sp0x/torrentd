package json

import (
	"testing"
)

func TestJSON(t *testing.T) {
	internal.SerializerTester(t, Serializer)
}
