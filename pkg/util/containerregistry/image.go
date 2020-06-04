package containerregistry

import (
	"bytes"
	"path"
)

var (
	registryDomain    string
	registryNamespace string
)

func Init(domain string, namespace string) {
	registryDomain = domain
	registryNamespace = namespace
}

type Image struct {
	Name string
	Tag  string
}

func (i Image) BaseName() string {
	b := new(bytes.Buffer)
	b.WriteString(i.Name)
	if i.Tag != "" {
		b.WriteString(":" + i.Tag)
	}
	return b.String()
}

func (i Image) FullName() string {
	return path.Join(registryDomain, registryNamespace, i.BaseName())
}

func GetImagePrefix(name string) string {
	return path.Join(registryDomain, registryNamespace, name)
}

func GetPrefix() string {
	return path.Join(registryDomain, registryNamespace)
}
