package metautil

import (
	"bytes"
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/template"
	"k8s.io/klog"
)

var MetaTemlate = `
apiVersion: devops.gostship.io/v1
kind: Cluster
metadata:
 name: host
 namespace: host
 annotations:
   kubesphere.io/description: "meta cluster"
 labels:
   cluster-role.kunkka.io/cluster-role: "meta"
spec:
 pause: false
 tenantID: kunkka
 displayName: host
 type: Baremetal
 version: v1.18.5
status:
 conditions:
 - lastProbeTime: "2020-08-03T12:22:14Z"
   reason: "Ready"
   status: "True"
   type: "Ready"
   lastTransitionTime: "2020-08-03T12:22:14Z"
   message: "Cluster is available now"
 version: "v1.18.5"
 nodeCount: 3
`

func BuildMetaObj() (*devopsv1.Cluster, error) {
	data, err := template.ParseString(MetaTemlate, "")
	var meta *devopsv1.Cluster
	var ok bool
	if err != nil {
		return nil, err
	}

	objs, err := k8sutil.LoadObjs(bytes.NewReader(data))
	if err != nil {
		klog.Errorf("bremetal load objs err: %v", err)
		return nil, err
	}
	for _, obj := range objs {
		if meta, ok = obj.(*devopsv1.Cluster); ok {
			break
		}
	}
	return meta, nil
}
