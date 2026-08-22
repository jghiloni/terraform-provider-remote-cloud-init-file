package provider_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v4"

	"github.com/jghiloni/terraform-provider-remote-cloud-init-file/internal/provider"
)

var _ = Describe("WriteFile", func() {
	var wfs = []string{`encoding: b64
path: /testb64
owner: "root:root"
permissions: "0640"
defer: true
content: |
  YjY0IHRlc3Q=`,
  	`encoding: gzip
path: /testgzip
permissions: "0700"
content: !!binary |
  H4sIAF4siWoAA0uvyixQKEktLgEAeP9tzgkAAAA=`,
		`path: /test
content: "plain test"`,
		`path: /testempty`,
	}

	DescribeTable("Unmarshal", func(yml string, encoding string, path string, owner string, permissions string, deferred bool, content string) {
		var wf provider.WriteFile
		Expect(yaml.NewDecoder(strings.NewReader(yml)).Decode(&wf)).NotTo(HaveOccurred())
		Expect(wf.Encoding).To(Equal(encoding))
		Expect(wf.Path).To(Equal(path))
		Expect(wf.Defer).To(Equal(deferred))
		Expect(wf.Owner).To(Equal(owner))
		Expect(wf.Permissions).To(Equal(permissions))
		Expect(wf.Content.Equals(content)).To(BeTrue())
	}, 
	Entry("b64", wfs[0], "b64", "/testb64", "root:root", "0640", true, "b64 test"), 
	Entry("gzip", wfs[1], "gzip", "/testgzip", "", "0700", false, "gzip test"),
	Entry("plain", wfs[2], "", "/test", "", "", false, "plain test"),
	Entry("empty", wfs[3], "", "/testempty", "", "", false, ""),
	)
	DescribeTable("Unmarshal", func(yml string, encoding string, path string, owner string, permissions string, deferred bool, content string) {
		wf := provider.WriteFile{
			Encoding: encoding,
			Path: path,
			Owner: owner,
			Permissions: permissions,
			Defer: deferred,
		}

		wf.SetContents(content)

		_, err := yaml.Marshal(wf)
		Expect(err).NotTo(HaveOccurred())

		var wf2 provider.WriteFile
		Expect(yaml.Unmarshal([]byte(yml), &wf2)).NotTo(HaveOccurred())
		Expect(wf).To(Equal(wf2))
	}, 
	Entry("b64", wfs[0], "b64", "/testb64", "root:root", "0640", true, "b64 test"), 
	Entry("gzip", wfs[1], "gzip", "/testgzip", "", "0700", false, "gzip test"),
	Entry("plain", wfs[2], "", "/test", "", "", false, "plain test"),
	Entry("empty", wfs[3], "", "/testempty", "", "", false, ""),
	)
})
