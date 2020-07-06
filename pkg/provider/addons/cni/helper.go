package cni

import (
	"bytes"

	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/provider"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/template"
)

const (
	initShellTemplate = `
#!/usr/bin/env bash

set -xeuo pipefail

function Bridge_network(){
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
 
cat << EOF | tee /etc/sysconfig/network-scripts/ifcfg-eth1
#ifcfg-eth1
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
systemctl restart NetworkManager
}
`

	hostLocalTemplate = `
{
 "cniVersion": "{{ default "0.3.1" .CniVersion}}",
 "name": "dke-cni",
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
     "subnet": "{{ .SubnetCidr}}",
     "rangeStart": "{{ .StartIP }}",
     "rangeEnd": "{{ .EndIP }}",
     "gateway": "{{ .Gw }}"
    }
   ]
  ],
  "routes": [
   {
    "dst": "0.0.0.0/0"
   },
   {
    "dst": "{{ .RouterDst }}",
    "gw": "{{ .RouterGw }}"
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

type Option struct {
	CniVersion string
	SubnetCidr string
	StartIP    string
	EndIP      string
	Gw         string
	RouterDst  string
	RouterGw   string
}

func Install(s ssh.Interface, c *provider.Cluster) error {
	opt := &Option{
		SubnetCidr: "10.49.255.0/24",
		StartIP:    "10.49.255.1",
		EndIP:      "10.49.255.40",
		Gw:         "10.49.255.254",
		RouterDst:  "10.27.248.0/24",
		RouterGw:   "10.28.252.241",
	}
	localByte, err := template.ParseString(hostLocalTemplate, opt)
	if err != nil {
		return err
	}

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
