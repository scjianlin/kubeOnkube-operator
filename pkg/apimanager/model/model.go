package model

import (
	v1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// 机柜CIDR
type Rack struct {
	ID           string           `json:"id"`
	RackCidr     string           `json:"rackCidr"`
	RackCidrGw   string           `json:"rackCidrGw"`
	ProviderCidr string           `json:"providerCidr"`
	RackTag      string           `json:"rackTag"`
	IsMaster     int              `json:"isMaster"` //值为0表示False,值为1表示True
	HostAddr     []*HostAddr      `json:"hostAddr"`
	PodCidr      []*v1.ClusterCni `json:"podCidr"`
	PodNum       int              `json:"podNum"`
}

// 主机地址
type HostAddr struct {
	ID        string   `json:"id"`
	IPADDR    string   `json:"ipAddr"`
	NetMask   string   `json:"netMask"`
	GateWay   string   `json:"gateWay"`
	DnsServer []string `json:"dnsServer"`
	UseState  int      `json:"useState"` //值0表示未使用,1表示已经使用
	IsMeta    int      `json:"isMeta"`   //值0表示不是meta集群地址,1表示是meta集群的节点地址
}

// POD 地址段
type PodAddr struct {
	ID           string `json:"id"`
	RangeStart   string `json:"rangeStart"`
	RangeEnd     string `json:"rangeEnd"`
	DefaultRoute string `json:"defaultRoute"`
	UseState     int    `json:"useState"` //值0表示未使用,1表示已经使用
}

//
type PodAddrList struct {
	ID           string `json:"id"`
	RangeStart   string `json:"rangeStart"`
	RangeEnd     string `json:"rangeEnd"`
	DefaultRoute string `json:"defaultRoute"`
	RackCidr     string `json:"rackCidr"`
	RackTag      string `json:"rackTag"`
}

// cluster version
type ClusterVersion struct {
	ID            string `json:"id"`
	MasterVersion string `json:"masterVersion"`
	DockerVersion string `json:"dockerVersion"`
}

// Add Cluster struct
type AddCluster struct {
	ClusterName    string   `json:"clusterName"`
	ClusterType    string   `json:"clusterType"`
	ClusterRack    []string `json:"clusterRack"`
	ClusterIP      []string `json:"clusterIp"`
	UserName       string   `json:"userName"`
	Password       string   `json:"passWord"`
	ClusterVersion string   `json:"clusterVersion"`
	DockerVersion  string   `json:"dockerVersion"`
	CustomScript   string   `json:"customScript"`
	Description    string   `json:"description"`
	ClusterGroup   string   `json:"clusterGroup"`
	PodPool        []string `json:"podPool"`
}

type CniOption struct {
	Racks   string         `json:"racks"`
	Machine string         `json:"machine"`
	Cni     *v1.ClusterCni `json:"cni"`
}

type ClusterNode struct {
	AddressList   []string `json:"addressList"`
	ClusterName   string   `json:"clusterName"`
	CustomScript  string   `json:"customScript"`
	DockerVersion string   `json:"dockerVersion"`
	NodeRack      []string `json:"nodeRack"`
	NodeVersion   string   `json:"nodeVersion"`
	Password      string   `json:"password"`
	PodPool       []string `json:"podPool"`
	UserName      string   `json:"userName"`
}

// cluster condition
type RuntimeCondition struct {
	Type   string             `json:"type"`
	Name   string             `json:"name"`
	Status v1.ConditionStatus `json:"status"`
	Time   metav1.Time        `json:"time"`
}

// cluster role model
type ClusterRole struct {
	Metadata Metadata    `json:"metadata"`
	Rules    interface{} `json:"rules"`
}
type Labels struct {
	IamKubesphereIoRoleTemplate string `json:"iam.kubesphere.io/role-template"`
}
type Annotations struct {
	IamKubesphereIoModule                       string `json:"iam.kubesphere.io/module"`
	IamKubesphereIoRoleTemplateRules            string `json:"iam.kubesphere.io/role-template-rules"`
	KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
	KubesphereIoAliasName                       string `json:"kubesphere.io/alias-name"`
}
type Metadata struct {
	Name              string      `json:"name"`
	SelfLink          string      `json:"selfLink"`
	UID               string      `json:"uid"`
	ResourceVersion   string      `json:"resourceVersion"`
	CreationTimestamp time.Time   `json:"creationTimestamp"`
	Labels            Labels      `json:"labels"`
	Annotations       Annotations `json:"annotations"`
}
