package authutil

import (
	"encoding/json"
	"github.com/gostship/kunkka/pkg/apimanager/model/auth"
)

var roleTemplate = `
{
 "items": [
  {
   "kind": "GlobalRole",
   "apiVersion": "iam.kubesphere.io/v1alpha2",
   "metadata": {
    "name": "admin",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/globalroles/admin",
    "uid": "735536fc-bd71-46b1-a777-e4d2d68186a7",
    "resourceVersion": "330428436",
    "generation": 1,
    "creationTimestamp": "2020-08-04T02:14:09Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/aggregation-roles": "[\"role-template-manage-roles\",\"role-template-view-users\",\"role-template-view-roles\",\"role-template-view-users\",\"role-template-view-roles\",\"role-template-manage-workspaces\",\"role-template-view-workspaces\",\"role-template-view-users\",\"role-template-view-roles\",\"role-template-manage-users\",\"role-template-view-users\",\"role-template-view-roles\",\"role-template-manage-platform-settings\",\"role-template-manage-clusters\",\"role-template-view-clusters\"]",
     "kubesphere.io/alias-name": "admin",
     "kubesphere.io/creator": "admin"
    }
   },
   "rules": [
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "globalroles"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "iam.kubesphere.io"
     ],
     "resources": [
      "globalroles"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "iam.kubesphere.io"
     ],
     "resources": [
      "globalroles"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "abnormalworkloads",
      "quotas",
      "workloads",
      "volumesnapshots",
      "dashboards",
      "configmaps",
      "endpoints",
      "events",
      "limitranges",
      "namespaces",
      "persistentvolumeclaims",
      "podtemplates",
      "replicationcontrollers",
      "resourcequotas",
      "secrets",
      "serviceaccounts",
      "services",
      "applications",
      "controllerrevisions",
      "deployments",
      "replicasets",
      "statefulsets",
      "daemonsets",
      "meshpolicies",
      "cronjobs",
      "jobs",
      "devopsprojects",
      "devops",
      "pipelines",
      "pipelines/runs",
      "pipelines/branches",
      "pipelines/checkScriptCompile",
      "pipelines/consolelog",
      "pipelines/scan",
      "pipelines/sonarstatus",
      "checkCron",
      "credentials",
      "credentials/usage",
      "s2ibinaries",
      "s2ibinaries/file",
      "s2ibuilders",
      "s2ibuildertemplates",
      "s2iruns",
      "events",
      "ingresses",
      "router",
      "filters",
      "pods",
      "pods/log",
      "namespacenetworkpolicies",
      "workspacenetworkpolicies",
      "networkpolicies",
      "podsecuritypolicies",
      "rolebindings",
      "roles",
      "members",
      "servicepolicies",
      "federatedapplications",
      "federatedconfigmaps",
      "federateddeployments",
      "federatedingresses",
      "federatedjobs",
      "federatedlimitranges",
      "federatednamespaces",
      "federatedpersistentvolumeclaims",
      "federatedreplicasets",
      "federatedsecrets",
      "federatedserviceaccounts",
      "federatedservices",
      "federatedservicestatuses",
      "federatedstatefulsets",
      "federatedworkspaces",
      "workspaces",
      "workspaceroles",
      "workspacemembers",
      "workspacemembers/namespaces",
      "workspacemembers/devops",
      "workspacerolebindings",
      "repos",
      "repos/action",
      "repos/events",
      "apps",
      "apps/versions",
      "categories",
      "apps/audits"
     ]
    },
    {
     "verbs": [
      "list"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "clusters"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "abnormalworkloads",
      "quotas",
      "workloads",
      "volumesnapshots",
      "dashboards",
      "configmaps",
      "endpoints",
      "events",
      "limitranges",
      "namespaces",
      "persistentvolumeclaims",
      "pods",
      "podtemplates",
      "replicationcontrollers",
      "resourcequotas",
      "secrets",
      "serviceaccounts",
      "services",
      "applications",
      "controllerrevisions",
      "deployments",
      "replicasets",
      "statefulsets",
      "daemonsets",
      "meshpolicies",
      "cronjobs",
      "jobs",
      "devopsprojects",
      "devops",
      "pipelines",
      "pipelines/runs",
      "pipelines/branches",
      "pipelines/checkScriptCompile",
      "pipelines/consolelog",
      "pipelines/scan",
      "pipelines/sonarstatus",
      "checkCron",
      "credentials",
      "credentials/usage",
      "s2ibinaries",
      "s2ibinaries/file",
      "s2ibuilders",
      "s2ibuildertemplates",
      "s2iruns",
      "events",
      "ingresses",
      "router",
      "filters",
      "pods",
      "pods/log",
      "namespacenetworkpolicies",
      "workspacenetworkpolicies",
      "networkpolicies",
      "podsecuritypolicies",
      "rolebindings",
      "roles",
      "members",
      "servicepolicies",
      "federatedconfigmaps",
      "federateddeployments",
      "federatedingresses",
      "federatedjobs",
      "federatedlimitranges",
      "federatednamespaces",
      "federatedpersistentvolumeclaims",
      "federatedreplicasets",
      "federatedsecrets",
      "federatedserviceaccounts",
      "federatedservices",
      "federatedservicestatuses",
      "federatedstatefulsets",
      "federatedworkspaces",
      "workspaces",
      "workspaceroles",
      "workspacemembers",
      "workspacemembers/namespaces",
      "workspacemembers/devops",
      "workspacerolebindings"
     ]
    },
    {
     "verbs": [
      "get",
      "list"
     ],
     "apiGroups": [
      "openpitrix.io"
     ],
     "resources": [
      "repos",
      "repos/action",
      "repos/events",
      "apps",
      "apps/versions",
      "categories",
      "apps/audits"
     ]
    },
    {
     "verbs": [
      "list"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "clusters",
      "cluster"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "iam.kubesphere.io"
     ],
     "resources": [
      "globalroles"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "iam.kubesphere.io"
     ],
     "resources": [
      "globalroles"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "logging.kubesphere.io"
     ],
     "resources": [
      "*"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "*"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "nonResourceURLs": [
      "*"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "*"
     ]
    },
    {
     "verbs": [
      "GET"
     ],
     "nonResourceURLs": [
      "*"
     ]
    }
   ]
  },
  {
   "kind": "GlobalRole",
   "apiVersion": "iam.kubesphere.io/v1alpha2",
   "metadata": {
    "name": "workspaces-manager",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/globalroles/workspaces-manager",
    "uid": "345ca38f-a0ed-48ca-8185-c4418979b5bb",
    "resourceVersion": "321212500",
    "generation": 2,
    "creationTimestamp": "2020-08-03T12:20:17Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/aggregation-roles": "[\"role-template-view-workspaces\",\"role-template-manage-workspaces\",\"role-template-view-users\"]",
     "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"iam.kubesphere.io/v1alpha2\",\"kind\":\"GlobalRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/aggregation-roles\":\"[\\\"role-template-view-workspaces\\\",\\\"role-template-manage-workspaces\\\",\\\"role-template-view-users\\\"]\",\"kubesphere.io/creator\":\"admin\"},\"name\":\"workspaces-manager\"},\"rules\":[{\"apiGroups\":[\"*\"],\"resources\":[\"abnormalworkloads\",\"quotas\",\"workloads\",\"volumesnapshots\",\"dashboards\",\"configmaps\",\"endpoints\",\"events\",\"limitranges\",\"namespaces\",\"persistentvolumeclaims\",\"pods\",\"podtemplates\",\"replicationcontrollers\",\"resourcequotas\",\"secrets\",\"serviceaccounts\",\"services\",\"applications\",\"controllerrevisions\",\"deployments\",\"replicasets\",\"statefulsets\",\"daemonsets\",\"meshpolicies\",\"cronjobs\",\"jobs\",\"devopsprojects\",\"devops\",\"pipelines\",\"pipelines/runs\",\"pipelines/branches\",\"pipelines/checkScriptCompile\",\"pipelines/consolelog\",\"pipelines/scan\",\"pipelines/sonarstatus\",\"checkCron\",\"credentials\",\"credentials/usage\",\"s2ibinaries\",\"s2ibinaries/file\",\"s2ibuilders\",\"s2ibuildertemplates\",\"s2iruns\",\"events\",\"ingresses\",\"router\",\"filters\",\"pods\",\"pods/log\",\"namespacenetworkpolicies\",\"workspacenetworkpolicies\",\"networkpolicies\",\"podsecuritypolicies\",\"rolebindings\",\"roles\",\"members\",\"servicepolicies\",\"federatedapplications\",\"federatedconfigmaps\",\"federateddeployments\",\"federatedingresses\",\"federatedjobs\",\"federatedlimitranges\",\"federatednamespaces\",\"federatedpersistentvolumeclaims\",\"federatedreplicasets\",\"federatedsecrets\",\"federatedserviceaccounts\",\"federatedservices\",\"federatedservicestatuses\",\"federatedstatefulsets\",\"federatedworkspaces\",\"workspaces\",\"workspaceroles\",\"workspacemembers\",\"workspacemembers/namespaces\",\"workspacemembers/devops\",\"workspacerolebindings\"],\"verbs\":[\"*\"]},{\"apiGroups\":[\"*\"],\"resources\":[\"users\",\"users/loginrecords\"],\"verbs\":[\"get\",\"list\",\"watch\"]},{\"apiGroups\":[\"openpitrix.io\"],\"resources\":[\"repos\",\"apps\",\"apps/versions\",\"categories\",\"apps/audits\"],\"verbs\":[\"*\"]},{\"apiGroups\":[\"*\"],\"resources\":[\"clusters\",\"cluster\"],\"verbs\":[\"list\"]}]}\n",
     "kubesphere.io/creator": "admin"
    }
   },
   "rules": [
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "abnormalworkloads",
      "quotas",
      "workloads",
      "volumesnapshots",
      "dashboards",
      "configmaps",
      "endpoints",
      "events",
      "limitranges",
      "namespaces",
      "persistentvolumeclaims",
      "pods",
      "podtemplates",
      "replicationcontrollers",
      "resourcequotas",
      "secrets",
      "serviceaccounts",
      "services",
      "applications",
      "controllerrevisions",
      "deployments",
      "replicasets",
      "statefulsets",
      "daemonsets",
      "meshpolicies",
      "cronjobs",
      "jobs",
      "devopsprojects",
      "devops",
      "pipelines",
      "pipelines/runs",
      "pipelines/branches",
      "pipelines/checkScriptCompile",
      "pipelines/consolelog",
      "pipelines/scan",
      "pipelines/sonarstatus",
      "checkCron",
      "credentials",
      "credentials/usage",
      "s2ibinaries",
      "s2ibinaries/file",
      "s2ibuilders",
      "s2ibuildertemplates",
      "s2iruns",
      "events",
      "ingresses",
      "router",
      "filters",
      "pods",
      "pods/log",
      "namespacenetworkpolicies",
      "workspacenetworkpolicies",
      "networkpolicies",
      "podsecuritypolicies",
      "rolebindings",
      "roles",
      "members",
      "servicepolicies",
      "federatedapplications",
      "federatedconfigmaps",
      "federateddeployments",
      "federatedingresses",
      "federatedjobs",
      "federatedlimitranges",
      "federatednamespaces",
      "federatedpersistentvolumeclaims",
      "federatedreplicasets",
      "federatedsecrets",
      "federatedserviceaccounts",
      "federatedservices",
      "federatedservicestatuses",
      "federatedstatefulsets",
      "federatedworkspaces",
      "workspaces",
      "workspaceroles",
      "workspacemembers",
      "workspacemembers/namespaces",
      "workspacemembers/devops",
      "workspacerolebindings"
     ]
    },
    {
     "verbs": [
      "get",
      "list",
      "watch"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "openpitrix.io"
     ],
     "resources": [
      "repos",
      "apps",
      "apps/versions",
      "categories",
      "apps/audits"
     ]
    },
    {
     "verbs": [
      "list"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "clusters",
      "cluster"
     ]
    }
   ]
  },
  {
   "kind": "GlobalRole",
   "apiVersion": "iam.kubesphere.io/v1alpha2",
   "metadata": {
    "name": "users-manager",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/globalroles/users-manager",
    "uid": "5fe9e61b-b168-4cc8-b8b1-92f39dd07710",
    "resourceVersion": "321212498",
    "generation": 2,
    "creationTimestamp": "2020-08-03T12:20:17Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/aggregation-roles": "[\"role-template-view-users\",\"role-template-manage-users\",\"role-template-view-roles\",\"role-template-manage-roles\"]",
     "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"iam.kubesphere.io/v1alpha2\",\"kind\":\"GlobalRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/aggregation-roles\":\"[\\\"role-template-view-users\\\",\\\"role-template-manage-users\\\",\\\"role-template-view-roles\\\",\\\"role-template-manage-roles\\\"]\",\"kubesphere.io/creator\":\"admin\"},\"name\":\"users-manager\"},\"rules\":[{\"apiGroups\":[\"*\"],\"resources\":[\"users\",\"users/loginrecords\",\"globalroles\"],\"verbs\":[\"*\"]}]}\n",
     "kubesphere.io/creator": "admin"
    }
   },
   "rules": [
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "users",
      "users/loginrecords",
      "globalroles"
     ]
    }
   ]
  },
  {
   "kind": "GlobalRole",
   "apiVersion": "iam.kubesphere.io/v1alpha2",
   "metadata": {
    "name": "platform-regular",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/globalroles/platform-regular",
    "uid": "70458d7c-0f9c-4d78-b974-82483c7f5c3b",
    "resourceVersion": "320937533",
    "generation": 1,
    "creationTimestamp": "2020-08-03T12:20:17Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/aggregation-roles": "[\"role-template-view-app-templates\"]",
     "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"iam.kubesphere.io/v1alpha2\",\"kind\":\"GlobalRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/aggregation-roles\":\"[\\\"role-template-view-app-templates\\\"]\",\"kubesphere.io/creator\":\"admin\"},\"name\":\"platform-regular\"},\"rules\":[]}\n",
     "kubesphere.io/creator": "admin"
    }
   },
   "rules": []
  },
  {
   "kind": "GlobalRole",
   "apiVersion": "iam.kubesphere.io/v1alpha2",
   "metadata": {
    "name": "platform-admin",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/globalroles/platform-admin",
    "uid": "abc68eb0-3d30-466a-af88-2d80ec8e474f",
    "resourceVersion": "320937708",
    "generation": 1,
    "creationTimestamp": "2020-08-03T12:20:17Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/aggregation-roles": "[\"role-template-manage-clusters\",\"role-template-view-clusters\",\"role-template-view-roles\",\"role-template-manage-roles\",\"role-template-view-roles\",\"role-template-view-workspaces\",\"role-template-manage-workspaces\",\"role-template-manage-users\",\"role-template-view-roles\",\"role-template-view-users\",\"role-template-manage-app-templates\",\"role-template-view-app-templates\",\"role-template-manage-platform-settings\"]",
     "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"iam.kubesphere.io/v1alpha2\",\"kind\":\"GlobalRole\",\"metadata\":{\"annotations\":{\"iam.kubesphere.io/aggregation-roles\":\"[\\\"role-template-manage-clusters\\\",\\\"role-template-view-clusters\\\",\\\"role-template-view-roles\\\",\\\"role-template-manage-roles\\\",\\\"role-template-view-roles\\\",\\\"role-template-view-workspaces\\\",\\\"role-template-manage-workspaces\\\",\\\"role-template-manage-users\\\",\\\"role-template-view-roles\\\",\\\"role-template-view-users\\\",\\\"role-template-manage-app-templates\\\",\\\"role-template-view-app-templates\\\",\\\"role-template-manage-platform-settings\\\"]\",\"kubesphere.io/creator\":\"admin\"},\"name\":\"platform-admin\"},\"rules\":[{\"apiGroups\":[\"*\"],\"resources\":[\"*\"],\"verbs\":[\"*\"]},{\"nonResourceURLs\":[\"*\"],\"verbs\":[\"*\"]}]}\n",
     "kubesphere.io/creator": "admin"
    }
   },
   "rules": [
    {
     "verbs": [
      "*"
     ],
     "apiGroups": [
      "*"
     ],
     "resources": [
      "*"
     ]
    },
    {
     "verbs": [
      "*"
     ],
     "nonResourceURLs": [
      "*"
     ]
    }
   ]
  }
 ],
 "totalItems": 5
}
`

func BuildGlobalRole() (*auth.GlobalRole, error) {
	obj := &auth.GlobalRole{}
	err := json.Unmarshal([]byte(roleTemplate), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
