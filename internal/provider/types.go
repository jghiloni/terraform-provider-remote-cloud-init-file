package provider

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v4"
)

type WriteFile struct {
	Path        string
	Encoding    string
	Owner       string
	Permissions string
	Defer       bool
	Content     writeFileContents
}

func (w *WriteFile) SetContents(s string) {
	w.Content = writeFileContents{
		content:  s,
		encoding: w.Encoding,
	}
}

type writeFileInternal struct {
	Path        string            `yaml:"path"`
	Encoding    string            `yaml:"encoding,omitempty"`
	Owner       string            `yaml:"owner,omitempty"`
	Permissions string            `yaml:"permissions,omitempty"`
	Defer       bool              `yaml:"defer,omitempty"`
	Content     writeFileContents `yaml:"content,omitempty"`
}

func (w *WriteFile) MarshalYAML() (any, error) {
	i := writeFileInternal{
		Encoding:    w.Encoding,
		Path:        w.Path,
		Owner:       w.Owner,
		Permissions: w.Permissions,
		Defer:       w.Defer,
		Content:     w.Content,
	}

	i.Content.encoding = i.Encoding

	node := new(yaml.Node)
	if err := node.Dump(i); err != nil {
		return nil, err
	}

	return node, nil
}

func (w *WriteFile) UnmarshalYAML(node *yaml.Node) error {
	var i writeFileInternal

	err := node.Load(&i)
	if err != nil {
		return err
	}

	if i.Encoding == "b64" {
		v, err := base64.StdEncoding.DecodeString(i.Content.content)
		if err != nil {
			return err
		}

		i.Content.content = string(v)
		i.Content.encoding = "b64"
	}

	*w = WriteFile(i)

	return nil
}

type writeFileContents struct {
	content  string
	encoding string
}

func (w *writeFileContents) Equals(s string) bool {
	if w == nil && s == "" {
		return true
	}
	return w.content == s
}

func (w *writeFileContents) UnmarshalYAML(node *yaml.Node) error {
	u := writeFileContents{}
	if node.Tag == "!!binary" {
		v, err := base64.StdEncoding.DecodeString(node.Value)
		if err != nil {
			return err
		}

		r, err := gzip.NewReader(bytes.NewBuffer(v))
		if err != nil {
			return err
		}

		b := new(strings.Builder)
		_, err = io.Copy(b, r)
		if err != nil {
			return err
		}

		u.content = b.String()
		u.encoding = "gzip"
		*w = u
		return nil
	}

	u.content = node.Value
	*w = u
	return nil
}

func (w *writeFileContents) MarshalYAML() (any, error) {
	encoded := ""
	tag := ""
	switch w.encoding {
	case "b64":
		encoded = base64.StdEncoding.EncodeToString([]byte(w.content))
	case "gzip":
		tag = "!!binary"
		buf := new(bytes.Buffer)

		encoder := gzip.NewWriter(buf)
		defer encoder.Close()

		_, err := encoder.Write([]byte(w.content))
		if err != nil {
			return nil, fmt.Errorf("could not gzip contents: %w", err)
		}
		encoder.Close()

		encoded = base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	node := new(yaml.Node)
	if err := node.Dump(encoded); err != nil {
		return nil, err
	}
	node.Tag = tag

	return node, nil
}
