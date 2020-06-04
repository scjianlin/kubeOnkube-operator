package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/jinzhu/configor"
)

func New(filename string) (*Config, error) {
	config := &Config{}
	if err := configor.Load(config, filename); err != nil {
		return nil, err
	}

	s := strings.Split(config.Registry.Prefix, "/")
	if len(s) != 2 {
		return nil, errors.New("invalid registry prefix")
	}
	config.Registry.Domain = s[0]
	config.Registry.Namespace = s[1]

	return config, nil
}

type Config struct {
	Registry Registry `yaml:"registry"`
	Audit    Audit    `yaml:"audit"`
	Feature  Feature  `yaml:"feature"`
}

func (c *Config) Save(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	y := yaml.NewEncoder(f)
	return y.Encode(c)
}

type Registry struct {
	Prefix    string `yaml:"prefix"`
	IP        string `yaml:"ip"`
	Domain    string `yaml:"-"`
	Namespace string `yaml:"-"`
}

func (r *Registry) NeedSetHosts() bool {
	return r.IP != ""
}

type Audit struct {
	Address string `yaml:"address"`
}

type Feature struct {
	SkipConditions []string `yaml:"skipConditions"`
}
