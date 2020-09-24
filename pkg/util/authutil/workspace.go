package authutil

import (
	"encoding/json"
	"errors"
	"github.com/ghodss/yaml"
)

type WorkSpace struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Generation int    `json:"generation"`
		Name       string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Placement struct {
			ClusterSelector struct {
			} `json:"clusterSelector"`
		} `json:"placement"`
		Template struct {
			Spec struct {
				Manager          string `json:"manager"`
				NetworkIsolation bool   `json:"networkIsolation"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

var templates = `
apiVersion: tenant.gostship.io/v1alpha2
kind: WorkspaceTemplate
metadata:
  generation: 1
  name: system-workspace
spec:
  placement:
    clusterSelector: {}
  template:
    spec:
      manager: admin
      networkIsolation: false
`

func BuildWorkspaceTemplate() ([]*WorkSpace, error) {
	obj := &WorkSpace{}

	objs := []*WorkSpace{}
	// 将yaml转换为json
	yamlToSpace, err := yaml.YAMLToJSON([]byte(templates))
	if err != nil {
		return nil, errors.New("workspace yaml to json error")
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToSpace, obj)
	if err != nil {
		return nil, errors.New("build runtime obj workspace error.")
	}
	objs = append(objs, obj)
	return objs, err
}
