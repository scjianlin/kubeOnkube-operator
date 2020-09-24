package workspace

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceTemplateSpec `json:"spec,omitempty"`
}

type WorkspaceTemplateSpec struct {
	Template  Template   `json:"template"`
	Placement Placement  `json:"placement"`
	Overrides []Override `json:"overrides,omitempty"`
}

type Template struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceSpec `json:"spec"`
}

type Placement struct {
	Clusters        []Cluster        `json:"clusters,omitempty"`
	ClusterSelector *ClusterSelector `json:"clusterSelector,omitempty"`
}

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	Manager          string `json:"manager,omitempty"`
	NetworkIsolation *bool  `json:"networkIsolation,omitempty"`
}

type ClusterSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Cluster struct {
	Name string `json:"name"`
}

type Override struct {
	ClusterName      string            `json:"clusterName"`
	ClusterOverrides []ClusterOverride `json:"clusterOverrides"`
}

type ClusterOverride struct {
	Path  string               `json:"path"`
	Op    string               `json:"op,omitempty"`
	Value runtime.RawExtension `json:"value"`
}
