package authutil

import (
	"encoding/json"
	"github.com/gostship/kunkka/pkg/apimanager/model/auth"
)

var user = `
{
 "items": [
  {
   "metadata": {
    "name": "admin",
    "selfLink": "/apis/iam.kubesphere.io/v1alpha2/users/admin",
    "uid": "1cfa3dba-bf15-422b-b428-af20ca11bf31",
    "resourceVersion": "1577237139",
    "generation": 227,
    "creationTimestamp": "2020-08-03T12:20:16Z",
    "labels": {
     "kubefed.io/managed": "false"
    },
    "annotations": {
     "iam.kubesphere.io/globalrole": "platform-admin",
     "iam.kubesphere.io/password-encrypted": "true"
    },
    "finalizers": [
     "finalizers.kubesphere.io/users"
    ]
   },
   "spec": {
    "email": "admin@kubesphere.io",
    "lang": "zh"
   },
   "status": {
    "state": "Active",
    "lastTransitionTime": "2020-08-03T12:48:10Z",
    "lastLoginTime": "2020-09-24T06:24:46Z"
   }
  }
 ],
 "totalItems": 1
}
`

func BuildUserMap() (*auth.UserMap, error) {
	obj := &auth.UserMap{}

	err := json.Unmarshal([]byte(user), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
