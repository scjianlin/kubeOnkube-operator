package config

import (
	"bytes"
	"errors"
	"path"
	"strings"
)

type Config struct {
	Registry    Registry
	Audit       Audit
	Feature     Feature
	CustomeCert bool
}

type Registry struct {
	Prefix    string
	IP        string
	Domain    string
	Namespace string
}

type Audit struct {
	Address string
}

type Feature struct {
	SkipConditions []string
}

func NewDefaultConfig() (*Config, error) {
	config := &Config{
		Registry: Registry{
			Prefix: "registry.aliyuncs.com/google_containers",
		},
	}

	s := strings.Split(config.Registry.Prefix, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid registry prefix")
	}
	config.Registry.Domain = s[0]
	config.Registry.Namespace = s[1]
	config.CustomeCert = true
	return config, nil
}

func (r *Registry) NeedSetHosts() bool {
	return r.IP != ""
}

func (r *Registry) ImageFullName(Name, Tag string) string {
	b := new(bytes.Buffer)
	b.WriteString(Name)
	if Tag != "" {
		b.WriteString(":" + Tag)
	}

	return path.Join(r.Domain, r.Namespace, b.String())
}
