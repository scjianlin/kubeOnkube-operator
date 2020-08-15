package metautil

import (
	"bytes"
	"github.com/gostship/kunkka/pkg/apimanager/model"
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
   kunkka.io/description: "meta cluster"
 labels:
   cluster-role.kunkka.io/cluster-role: "meta"
   cluster.kunkka.io/group: "production"
spec:
 pause: false
 tenantID: kunkka
 displayName: host
 type: Baremetal
 version: v1.18.5
 machines:
  - ip: 10.248.224.183
    port: 22
    username: root
    password: "hNKKTFCAOp6r58A"
  - ip: 10.248.224.201
    port: 22
    username: root
    password: "hNKKTFCAOp6r58A"
  - ip: 10.248.224.199
    port: 22
    username: root
    password: "hNKKTFCAOp6r58A"
status:
 conditions:
 - lastProbeTime: "2020-08-03T12:22:14Z"
   reason: "Ready"
   status: "True"
   type: "Ready"
   lastTransitionTime: "2020-08-03T12:22:14Z"
   message: "Cluster is available now"
 version: "v1.18.5"
 phase: "Running"
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

func ConditionOfContains(cond1 []devopsv1.ClusterCondition, cond2 *model.ClusterCondition) *model.ClusterCondition {
	for _, con := range cond1 {
		if con.Type == cond2.Type {
			cond2.Status = con.Status
			cond2.Time = con.LastProbeTime
		}
	}
	return cond2
}
