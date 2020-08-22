package cni

import (
	"bytes"
	"fmt"
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"os"
	"strings"

	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/template"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	cniInitShell = `
#!/usr/bin/env bash

set -xeuo pipefail

#cni0
cat << EOF | tee /etc/sysconfig/network-scripts/ifcfg-cni0
TYPE=bridge
ONBOOT=yes
DEVICE=cni0
BOOTPROTO=static
IPV4_FAILURE_FATAL=no
NAME=cni0
BRIDGE_STP=yes
EOF

#!/usr/bin/env bash

set -xeuo pipefail

#cni0
cat << EOF | tee /etc/sysconfig/network-scripts/ifcfg-cni0
TYPE=bridge
ONBOOT=yes
DEVICE=cni0
BOOTPROTO=static
IPV4_FAILURE_FATAL=no
NAME=cni0
BRIDGE_STP=yes
EOF

egrep -i "IPADDR|PREFIX|NETMASK|GATEWAY" /etc/sysconfig/network-scripts/ifcfg-eth1 >> /etc/sysconfig/network-scripts/ifcfg-cni0
 
#ifcfg-eth1
cat << EOF | tee /etc/sysconfig/network-scripts/ifcfg-eth1
TYPE=Ethernet
PROXY_METHOD=none
BROWSER_ONLY=no
BOOTPROTO=none
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
NAME=eth1
DEVICE=eth1
ONBOOT=yes
BRIDGE=cni0
EOF

egrep -i "IPADDR|PREFIX|NETMASK|GATEWAY" /etc/sysconfig/network-scripts/ifcfg-eth1 >> /etc/sysconfig/network-scripts/ifcfg-cni0
 
#ifcfg-eth1
cat << EOF | tee /etc/sysconfig/network-scripts/ifcfg-eth1
TYPE=Ethernet
PROXY_METHOD=none
BROWSER_ONLY=no
BOOTPROTO=none
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
NAME=eth1
DEVICE=eth1
ONBOOT=yes
BRIDGE=cni0
EOF
`

	hostLocalTemplate = `
{
 "cniVersion": "{{ default "0.3.1" .CniVersion}}",
 "name": "k8s-cni",
 "type": "bridge",
 "bridge": "cni0",
 "forceAddress": false,
 "ipMasq": true,
 "hairpinMode": true,
 "ipam": {
  "type": "host-local",
  "ranges": [
   [
    {
     "subnet": "{{ .Subnet}}",
     "rangeStart": "{{ .RangeStart }}",
     "rangeEnd": "{{ .RangeEnd }}",
     "gateway": "{{ .Gateway }}"
    }
   ]
  ],
  "routes": [
   {
    "dst": "0.0.0.0/0"
   },
   {
    "dst": "{{ .Dst }}",
    "gw": "{{ .Gw }}"
   }
  ],
  "dataDir": "/opt/k8s/data/cni"
 }
}
`
	loopbackTemplate = `
{
 "cniVersion": "{{ default "0.3.1" .CniVersion}}",
 "name": "lo",
 "type": "loopback"
}
`
)

const (
	CniHostLocalConfig = "cni-host-local-config"
	Eth1CfgPath        = "/etc/sysconfig/network-scripts/ifcfg-eth1"
	Cni0CfgPath        = "/etc/sysconfig/network-scripts/ifcfg-cni0"
)

type Option struct {
	CniVersion string `json:"cniVersion,omitempty"`
	Subnet     string `json:"subnet,omitempty"`
	RangeStart string `json:"rangeStart,omitempty"`
	RangeEnd   string `json:"rangeEnd,omitempty"`
	Gateway    string `json:"gateway,omitempty"`
	Dst        string `json:"dst,omitempty"`
	Gw         string `json:"gw,omitempty"`
}

func ApplyEth(s ssh.Interface, c *common.Cluster) error {
	err := s.WriteFile(strings.NewReader(cniInitShell), constants.SystemInitCniFile)
	if err != nil {
		return err
	}

	if exist, _ := s.Exist(Cni0CfgPath); exist {
		klog.Warningf("node: %s file: %s always exist", s.HostIP(), Cni0CfgPath)
		return nil
	}

	if exist, _ := s.Exist(Eth1CfgPath); !exist {
		klog.Warningf("node: %s file: %s not exist", s.HostIP(), Eth1CfgPath)
		return nil
	}

	klog.Infof("node: %s start exec init eth ... ", s.HostIP())
	cmd := fmt.Sprintf("chmod a+x %s && %s", constants.SystemInitCniFile, constants.SystemInitCniFile)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return errors.Wrapf(err, "node: %s exec cmd: %s", s.HostIP(), cmd)
	}

	klog.Infof("node: %s restart network", s.HostIP())
	_, _ = s.CombinedOutput("systemctl restart network")
	return nil
}

func ApplyClusterCni(s ssh.Interface, c *common.Cluster, machine *devopsv1.ClusterMachine) error {
	//cluster := &devopsv1.Cluster{}
	//err := c.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Cluster.Namespace, Name: c.ClusterName}, cluster)
	//if err != nil {
	//	klog.Warningf("cluster: %s get cni cfgMap err: %v", c.Cluster.Name, err)
	//	return nil
	//}
	opt := &Option{
		Subnet:     machine.HostCni.Subnet,
		RangeEnd:   machine.HostCni.RangeEnd,
		RangeStart: machine.HostCni.RangeStart,
		Gateway:    machine.HostCni.GW,
		Dst:        machine.HostCni.DefaultRoute,
		Gw:         machine.IP,
	}
	//opt := &Option{
	//	Subnet:     "10.28.0.0/22",
	//	RangeEnd:   "10.28.0.1",
	//	RangeStart: "10.28.0.240",
	//	Gateway:    "10.28.3.254",
	//	Dst:        "10.28.247.0/22",
	//	Gw:         machine.IP,
	//}

	localByte, err := template.ParseString(hostLocalTemplate, opt)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(localByte), constants.CniHostLocalFile)
	if err != nil {
		return err
	}

	klog.Infof("build node: %s cni :%s", s.HostIP(), string(localByte))

	loopByte, err := template.ParseString(loopbackTemplate, opt)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(loopByte), constants.CniLoopBack)
	if err != nil {
		return err
	}

	return nil
}

func ApplyNodeCni(s ssh.Interface, c *common.Cluster, machine *devopsv1.Machine) error {
	//node := &devopsv1.Machine{}
	//err := c.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Cluster.Namespace, Name: machine.Name}, node)
	//if err != nil {
	//	klog.Warningf("cluster: %s get cni cfgMap err: %v", c.Cluster.Name, err)
	//	return nil
	//}

	opt := &Option{
		Subnet:     machine.Spec.Machine.HostCni.Subnet,
		RangeEnd:   machine.Spec.Machine.HostCni.RangeEnd,
		RangeStart: machine.Spec.Machine.HostCni.RangeStart,
		Gateway:    machine.Spec.Machine.HostCni.GW,
		Dst:        machine.Spec.Machine.HostCni.DefaultRoute,
		Gw:         machine.Name,
	}
	//opt := &Option{
	//	Subnet:     "10.28.0.0/22",
	//	RangeEnd:   "10.28.0.1",
	//	RangeStart: "10.28.0.240",
	//	Gateway:    "10.28.3.254",
	//	Dst:        "10.28.247.0/22",
	//	Gw:         machine.Name,

	localByte, err := template.ParseString(hostLocalTemplate, opt)
	if err != nil {
		return err
	}

	//}klog.Infof("build node: %s cni: %s", s.HostIP(), string(localByte))

	err = s.WriteFile(bytes.NewReader(localByte), constants.CniHostLocalFile)
	if err != nil {
		return err
	}

	loopByte, err := template.ParseString(loopbackTemplate, opt)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(loopByte), constants.CniLoopBack)
	if err != nil {
		return err
	}

	return nil
}
