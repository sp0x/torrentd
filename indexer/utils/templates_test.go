package utils

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestApplyTemplate(t *testing.T) {
	g := gomega.NewWithT(t)
	ctxt := make(map[string]string)
	ctxt["key"] = "value"
	output, err := ApplyTemplate("nm", "{{replace \"somewhere\" \"where\" .key }}", ctxt)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(output).To(gomega.Equal("somevalue"))
}
