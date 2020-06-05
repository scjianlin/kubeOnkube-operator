package system

const (
	initShellTemplate = `
#!/usr/bin/env bash

set -xeuo pipefail

function Firewalld_process() {
    systemctl status firewalld | grep running

    echo -e "\033[32;32m 关闭防火墙 \033[0m \n"
    systemctl stop firewalld && systemctl disable firewalld

    echo -e "\033[32;32m 关闭selinux \033[0m \n"
    setenforce 0
    sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
}

function Set_hostname(){
    grep {{ .HostName }} /etc/hostname && echo -e "\033[32;32m 主机名已设置，退出设置主机名步骤 \033[0m \n" && return
	hostname {{ .HostName }}
	echo "{{ .HostName }}" > /etc/hostname
	echo "{{ .HostIP }} {{ .HostName }}" >> /etc/hosts
}

function Install_depend_environment(){
    rpm -qa | grep nfs-utils &> /dev/null && echo -e "\033[32;32m 已完成依赖环境安装，退出依赖环境安装步骤 \033[0m \n" && return

    yum install -y nfs-utils curl yum-utils device-mapper-persistent-data lvm2 \
           net-tools conntrack-tools wget vim  ntpdate libseccomp libtool-ltdl telnet \
           ipvsadm tc ipset bridge-utils tree telnet wget net-tools bash-completion sysstat

    echo -e "\033[32;32m 升级Centos7系统内核到5版本，解决Docker-ce版本兼容问题\033[0m \n"
    rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org && \
    rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-3.el7.elrepo.noarch.rpm && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel repolist && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml.x86_64 && \
    yum remove -y kernel-tools-libs.x86_64 kernel-tools.x86_64 && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml-tools.x86_64 && \
    grub2-set-default 0

    cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
vm.swappiness=0 # 禁止使用 swap 空间，只有当系统 OOM 时才允许使用它
vm.overcommit_memory=1 # 不检查物理内存是否够用
vm.panic_on_oom=0 # 开启 OOM
fs.inotify.max_user_instances=8192
fs.inotify.max_user_watches=1048576
fs.file-max=52706963
fs.nr_open=52706963
net.ipv6.conf.all.disable_ipv6=1
net.netfilter.nf_conntrack_max=2310720
EOF
    sysctl -p /etc/sysctl.d/k8s.conf
    ls /proc/sys/net/bridge

    # modprobe 
    modprobe br_netfilter
    modprobe ip_vs
    modprobe ip_vs_rr
    modprobe ip_vs_wrr
    modprobe ip_vs_sh
    modprobe nf_conntrack
    modinfo nf_conntrack_ipv4 && modprobe nf_conntrack_ipv4 && export nf_conntrack_ipv4="nf_conntrack_ipv4"

    cat > /etc/sysconfig/modules/ipvs.modules <<EOF
#!/bin/bash
modprobe -- ip_vs
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- ip_vs_sh
modprobe -- nf_conntrack_ipv4
EOF

    cat > /etc/modules-load.d/ip_vs.conf <<EOF
ip_vs
ip_vs_rr
ip_vs_wrr
ip_vs_sh
nf_conntrack
$nf_conntrack_ipv4
EOF
}

function Install_docker(){
    rpm -qa | grep docker && echo -e "\033[32;32m 已安装docker，退出安装docker步骤 \033[0m \n" && return
    yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
    yum makecache fast
    yum -y install docker-ce-{{ .DockerVersion }} docker-ce-cli-{{ .DockerVersion }}

    cat <<EOF > /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver={{ .Cgroupdriver }}"],
  "exec-root": "",
  "graph": "/var/lib/docker",
  "group": "",
  "ip-forward": true,
  "ip-masq": false,
  "iptables": false,
  "ipv6": false,
  "labels": [],
  "live-restore": true,
  "log-driver": "json-file",
  "log-level": "warn",
  "log-opts": {
    "max-file": "10",
    "max-size": "100m"
  },
  "registry-mirrors": [
    "https://mirror.ccs.tencentyun.com",
    "https://4xr1qpsp.mirror.aliyuncs.com"
  ],
{{- if .InsecureRegistries }}
  "insecure-registries": [
    {{ .InsecureRegistries }}
  ],
{{- end}}
  "runtimes": {},
  "selinux-enabled": false,
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF
	systemctl enable docker.service
    systemctl daemon-reload
    systemctl restart docker
}

function Install_kubernetes_component(){
    rpm -qa | grep kubernetes && echo -e "\033[32;32m 已安装kubernetes组件，退出 \033[0m \n" && return
    cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
	yum makecache fast
    yum -y install kubelet-{{ .K8sVersion }} kubeadm-{{ .K8sVersion }} kubectl-{{ .K8sVersion }} kubernetes-cni
}

# 初始化顺序

Firewalld_process && \
HostName && \
Set_hostname && \
Install_depend_environment && \
Install_docker
`
)
