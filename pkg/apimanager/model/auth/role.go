package auth

import "time"

type GlobalRole struct {
	Items []struct {
		Kind       string `json:"kind"`
		APIVersion string `json:"apiVersion"`
		Metadata   struct {
			Name              string    `json:"name"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			Generation        int       `json:"generation"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				KubefedIoManaged string `json:"kubefed.io/managed"`
			} `json:"labels"`
			Annotations struct {
				IamKubesphereIoAggregationRoles string `json:"iam.kubesphere.io/aggregation-roles"`
				KubesphereIoAliasName           string `json:"kubesphere.io/alias-name"`
				KubesphereIoCreator             string `json:"kubesphere.io/creator"`
			} `json:"annotations"`
		} `json:"metadata"`
		Rules []struct {
			Verbs           []string `json:"verbs"`
			APIGroups       []string `json:"apiGroups,omitempty"`
			Resources       []string `json:"resources,omitempty"`
			NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
		} `json:"rules"`
	} `json:"items"`
	TotalItems int `json:"totalItems"`
}
