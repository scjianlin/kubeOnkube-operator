package metautil

import (
	"bytes"
	"encoding/json"
	"errors"
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

var ClusterRoleTemp = `
[
 {
  "metadata": {
   "name": "role-template-view-volumes",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-volumes",
   "uid": "90e91de6-1ee7-4629-9773-e599ad510d54",
   "resourceVersion": "320915085",
   "creationTimestamp": "2020-08-03T12:20:18Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Storage Management",
    "iam.kubesphere.io/role-template-rules": "{\"volumes\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Storage Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"volumes\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Volumes View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-volumes\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Volumes View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-volume-snapshots",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-volume-snapshots",
   "uid": "b95f0ae8-d560-476e-aafc-99d3597cf41a",
   "resourceVersion": "320915081",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-volumes\"]",
    "iam.kubesphere.io/module": "Storage Management",
    "iam.kubesphere.io/role-template-rules": "{\"volume-snapshots\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-volumes\\\"]\",\"iam.kubesphere.io/module\":\"Storage Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"volume-snapshots\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Volume Snapshots View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-volume-snapshots\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Volume Snapshots View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-storageclasses",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-storageclasses",
   "uid": "84b8f961-e65c-4545-b888-a47c6e704b11",
   "resourceVersion": "320915078",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-volumes\"]",
    "iam.kubesphere.io/module": "Storage Management",
    "iam.kubesphere.io/role-template-rules": "{\"storageclasses\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-volumes\\\"]\",\"iam.kubesphere.io/module\":\"Storage Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"storageclasses\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"StorageClasses View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-storageclasses\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "StorageClasses View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-roles",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-roles",
   "uid": "0098b66b-c951-47e2-b1b8-16b5357263c6",
   "resourceVersion": "320915074",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Access Control",
    "iam.kubesphere.io/role-template-rules": "{\"roles\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Access Control\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"roles\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Roles View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-roles\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Roles View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-projects",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-projects",
   "uid": "2b77f6de-8f77-4f04-b548-08e4fcac796e",
   "resourceVersion": "320915071",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Project Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"projects\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Project Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"projects\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Projects View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-projects\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Projects View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-nodes",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-nodes",
   "uid": "edce027c-f7eb-44e6-b62d-75dbbcd7856d",
   "resourceVersion": "320915068",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Cluster Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"nodes\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Cluster Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"nodes\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Nodes View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-nodes\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Nodes View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-network-policies",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-network-policies",
   "uid": "b2fda8bc-3edd-4db3-9371-bb0c973004a8",
   "resourceVersion": "320915066",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Network Management",
    "iam.kubesphere.io/role-template-rules": "{\"networkpolicies\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Network Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"networkpolicies\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Network Policies View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-network-policies\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Network Policies View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-members",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-members",
   "uid": "77116a98-87f6-43e3-b0a6-8c6f25f3cec3",
   "resourceVersion": "320915064",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Access Control",
    "iam.kubesphere.io/role-template-rules": "{\"members\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Access Control\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"members\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Members View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-members\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Members View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-cluster-monitoring",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-cluster-monitoring",
   "uid": "f9023159-7c46-4bd4-b566-2a2945e9b91a",
   "resourceVersion": "320915061",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Monitoring \u0026 Alerting",
    "iam.kubesphere.io/role-template-rules": "{\"monitoring\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Monitoring \\u0026 Alerting\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"monitoring\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Monitoring View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-cluster-monitoring\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Monitoring View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-app-workloads",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-app-workloads",
   "uid": "ccb9c267-3123-4c2f-8296-a70b7467ed7a",
   "resourceVersion": "320915059",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-projects\"]",
    "iam.kubesphere.io/module": "Project Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"deployments\": \"view\", \"statefulsets\": \"view\", \"daemonsets\": \"view\", \"jobs\": \"view\", \"cronjobs\": \"view\", \"pods\": \"view\", \"services\": \"view\", \"ingresses\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-projects\\\"]\",\"iam.kubesphere.io/module\":\"Project Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"deployments\\\": \\\"view\\\", \\\"statefulsets\\\": \\\"view\\\", \\\"daemonsets\\\": \\\"view\\\", \\\"jobs\\\": \\\"view\\\", \\\"cronjobs\\\": \\\"view\\\", \\\"pods\\\": \\\"view\\\", \\\"services\\\": \\\"view\\\", \\\"ingresses\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Application Workloads View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-app-workloads\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Application Workloads View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-alerting-policies",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-alerting-policies",
   "uid": "ea2d57c5-f85a-46ef-89d9-ae9f72aea97f",
   "resourceVersion": "320915058",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-alerting-messages\"]",
    "iam.kubesphere.io/module": "Monitoring \u0026 Alerting",
    "iam.kubesphere.io/role-template-rules": "{\"alert-policies\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-alerting-messages\\\"]\",\"iam.kubesphere.io/module\":\"Monitoring \\u0026 Alerting\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"alert-policies\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Alerting Policies View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-alerting-policies\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Alerting Policies View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-view-alerting-messages",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-view-alerting-messages",
   "uid": "f4c11f61-ecaa-4c00-b190-bc81538c2312",
   "resourceVersion": "320915055",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Monitoring \u0026 Alerting",
    "iam.kubesphere.io/role-template-rules": "{\"alert-messages\": \"view\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Monitoring \\u0026 Alerting\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"alert-messages\\\": \\\"view\\\"}\",\"kubesphere.io/alias-name\":\"Alerting Messages View\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-view-alerting-messages\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Alerting Messages View"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-volumes",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-volumes",
   "uid": "a2f2db4e-94b3-4dea-ab3d-e2d6a6fd6333",
   "resourceVersion": "320915052",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-volumes\", \"role-template-view-storageclasses\"]",
    "iam.kubesphere.io/module": "Storage Management",
    "iam.kubesphere.io/role-template-rules": "{\"volumes\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-volumes\\\", \\\"role-template-view-storageclasses\\\"]\",\"iam.kubesphere.io/module\":\"Storage Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"volumes\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Volumes Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-volumes\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Volumes Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-storageclasses",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-storageclasses",
   "uid": "78c93215-2e19-44e1-bb12-2ac685e9d9ef",
   "resourceVersion": "320915046",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-volumes\", \"role-template-view-storageclasses\"]",
    "iam.kubesphere.io/module": "Storage Management",
    "iam.kubesphere.io/role-template-rules": "{\"storageclasses\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-volumes\\\", \\\"role-template-view-storageclasses\\\"]\",\"iam.kubesphere.io/module\":\"Storage Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"storageclasses\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"StorageClasses Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-storageclasses\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "StorageClasses Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-roles",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-roles",
   "uid": "ecab86a0-2294-4681-b905-c650d68321dd",
   "resourceVersion": "320915043",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-roles\"]",
    "iam.kubesphere.io/module": "Access Control",
    "iam.kubesphere.io/role-template-rules": "{\"roles\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-roles\\\"]\",\"iam.kubesphere.io/module\":\"Access Control\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"roles\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Roles Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-roles\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Roles Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-projects",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-projects",
   "uid": "4ad8c280-7375-4eb2-8fc6-8e9b4c22e581",
   "resourceVersion": "320915039",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-projects\"]",
    "iam.kubesphere.io/module": "Project Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"projects\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-projects\\\"]\",\"iam.kubesphere.io/module\":\"Project Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"projects\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Projects Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-projects\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Projects Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-nodes",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-nodes",
   "uid": "eb41abfd-5327-4e72-bd0e-10a451ac6766",
   "resourceVersion": "320915038",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-nodes\"]",
    "iam.kubesphere.io/module": "Cluster Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"nodes\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-nodes\\\"]\",\"iam.kubesphere.io/module\":\"Cluster Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"nodes\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Nodes Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-nodes\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Nodes Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-network-policies",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-network-policies",
   "uid": "43e66b37-41eb-4d86-92dc-369be4a7d931",
   "resourceVersion": "320915036",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-network-policies\"]",
    "iam.kubesphere.io/module": "Network Management",
    "iam.kubesphere.io/role-template-rules": "{\"networkpolicies\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-network-policies\\\"]\",\"iam.kubesphere.io/module\":\"Network Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"networkpolicies\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Network Policies Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-network-policies\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Network Policies Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-members",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-members",
   "uid": "5395a1db-a352-4c62-90fa-c64063be97cc",
   "resourceVersion": "320915033",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-roles\", \"role-template-view-members\"]",
    "iam.kubesphere.io/module": "Access Control",
    "iam.kubesphere.io/role-template-rules": "{\"members\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-roles\\\", \\\"role-template-view-members\\\"]\",\"iam.kubesphere.io/module\":\"Access Control\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"members\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Members Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-members\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Members Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-crds",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-crds",
   "uid": "857690ee-6abc-4f7b-9137-9d19b0b9aeab",
   "resourceVersion": "320915010",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Cluster Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"customresources\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Cluster Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"customresources\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"CRD Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-crds\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "CRD Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-components",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-components",
   "uid": "070493cd-1520-4628-a957-17b3bc03c1e0",
   "resourceVersion": "320915030",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Cluster Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"components\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Cluster Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"components\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Components Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-components\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Components Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-cluster-settings",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-cluster-settings",
   "uid": "ace15d0c-5f20-4bf4-8b76-15e231d1002e",
   "resourceVersion": "320915026",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/module": "Cluster Settings",
    "iam.kubesphere.io/role-template-rules": "{\"cluster-settings\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/module\":\"Cluster Settings\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"cluster-settings\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Cluster Settings\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-cluster-settings\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Cluster Settings"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-app-workloads",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-app-workloads",
   "uid": "23488d3b-de57-42ef-a913-9d2728397f48",
   "resourceVersion": "320915022",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-app-workloads\", \"role-template-view-projects\"]",
    "iam.kubesphere.io/module": "Project Resources Management",
    "iam.kubesphere.io/role-template-rules": "{\"deployments\": \"manage\", \"statefulsets\": \"manage\", \"daemonsets\": \"manage\", \"jobs\": \"manage\", \"cronjobs\": \"manage\", \"pods\": \"manage\", \"services\": \"manage\", \"ingresses\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-app-workloads\\\", \\\"role-template-view-projects\\\"]\",\"iam.kubesphere.io/module\":\"Project Resources Management\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"deployments\\\": \\\"manage\\\", \\\"statefulsets\\\": \\\"manage\\\", \\\"daemonsets\\\": \\\"manage\\\", \\\"jobs\\\": \\\"manage\\\", \\\"cronjobs\\\": \\\"manage\\\", \\\"pods\\\": \\\"manage\\\", \\\"services\\\": \\\"manage\\\", \\\"ingresses\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Application Workloads Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-app-workloads\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Application Workloads Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-alerting-policies",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-alerting-policies",
   "uid": "a0954642-efb8-4ed3-b6d8-c7233970b0ef",
   "resourceVersion": "320915020",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[\"role-template-view-alerting-policies\", \"role-template-view-alerting-messages\"]",
    "iam.kubesphere.io/module": "Monitoring \u0026 Alerting",
    "iam.kubesphere.io/role-template-rules": "{\"alert-policies\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[\\\"role-template-view-alerting-policies\\\", \\\"role-template-view-alerting-messages\\\"]\",\"iam.kubesphere.io/module\":\"Monitoring \\u0026 Alerting\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"alert-policies\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Alerting Policies Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-alerting-policies\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Alerting Policies Management"
   }
  },
  "rules": null
 },
 {
  "metadata": {
   "name": "role-template-manage-alerting-messages",
   "selfLink": "/apis/rbac.authorization.k8s.io/v1/clusterroles/role-template-manage-alerting-messages",
   "uid": "c652a846-3133-4935-9f2a-da4b69e2b50f",
   "resourceVersion": "320915016",
   "creationTimestamp": "2020-08-03T12:20:17Z",
   "labels": {
    "iam.kubesphere.io/role-template": "true"
   },
   "annotations": {
    "iam.kubesphere.io/dependencies": "[role-template-view-alerting-messages\"]",
    "iam.kubesphere.io/module": "Monitoring \u0026 Alerting",
    "iam.kubesphere.io/role-template-rules": "{\"alert-messages\": \"manage\"}",
    "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/dependencies\":\"[role-template-view-alerting-messages\\\"]\",\"iam.kubesphere.io/module\":\"Monitoring \\u0026 Alerting\",\"iam.kubesphere.io/role-template-rules\":\"{\\\"alert-messages\\\": \\\"manage\\\"}\",\"kubesphere.io/alias-name\":\"Alerting Messages Management\"},\"labels\":{\"iam.kubesphere.io/role-template\":\"true\"},\"name\":\"role-template-manage-alerting-messages\"},\"rules\":[]}\n",
    "kubesphere.io/alias-name": "Alerting Messages Management"
   }
  },
  "rules": null
 }
]
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

func BuildClusterRole() ([]*model.ClusterRole, error) {
	role := []*model.ClusterRole{}

	err := json.Unmarshal([]byte(ClusterRoleTemp), &role)
	if err != nil {
		return nil, errors.New("build cluster role error")
	}
	return role, nil
}

func StringofContains(tag string, tags []string) bool {
	for _, v := range tags {
		if v == tag {
			return true
		}
	}
	return false
}
