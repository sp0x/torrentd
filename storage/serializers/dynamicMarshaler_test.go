package serializers_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/bots"
	"github.com/sp0x/torrentd/storage/serializers"
	"github.com/sp0x/torrentd/storage/serializers/json"
)

func TestDynamicMarshaler_Unmarshal(t *testing.T) {
	g := gomega.NewWithT(t)
	m := serializers.NewDynamicMarshaler(&bots.Chat{}, json.Serializer)
	c := bots.Chat{ChatId: 11}
	data, _ := json.Serializer.Marshal(&c)
	result, _ := m.Unmarshal(data)
	g.Expect(result).ToNot(gomega.BeNil())
	g.Expect(result.(*bots.Chat).ChatId).To(gomega.BeEquivalentTo(11))
}
