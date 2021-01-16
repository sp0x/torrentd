package internal

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/storage/serializers"
)

type testingStruct struct {
	Name string
}

func SerializerTester(t *testing.T, serializer serializers.MarshalUnmarshaler) {
	g := NewWithT(t)
	inValue := &testingStruct{"test"}
	decodedValue := &testingStruct{}
	encoded, err := serializer.Marshal(inValue)
	if err != nil {
		t.Fatal(err)
	}
	// g.Expect(string(encoded)).To(Equal(`{"Name":"test"}`))
	err = serializer.Unmarshal(encoded, &decodedValue)
	if err != nil {
		t.Fatal(err)
	}
	g.Expect(decodedValue).To(Equal(inValue))
}
