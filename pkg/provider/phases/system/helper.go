package system

const (
	initShellTemplate = `
#!/usr/bin/env bash

set -xeuo pipefail

function Firewalld_process() {
    grep SELINUX=disabled /etc/selinux/config && echo -e "\033[32;32m 已关闭防火墙，退出防火墙设置 \033[0m \n" && return

    echo -e "\033[32;32m 关闭防火墙 \033[0m \n"
    systemctl stop firewalld && systemctl disable firewalld

    echo -e "\033[32;32m 关闭selinux \033[0m \n"
    setenforce 0
    sed -i 's/^SELINUX=.*/SELINUX=disabled/' /etc/selinux/config
    
    swapoff -a && sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
}

function Install_depend_environment(){
    if [ -f /etc/sysctl.d/k8s.conf ]; then
      echo -e "\033[32;32m 已完成依赖环境安装 \033[0m \n" 
      return
    fi

    echo -e "\033[32;32m 开始安装依赖环境包 \033[0m \n" 
    yum makecache fast
    yum install -y nfs-utils curl yum-utils device-mapper-persistent-data lvm2 \
           net-tools conntrack-tools wget vim  ntpdate libseccomp libtool-ltdl telnet \
           ipvsadm tc ipset bridge-utils tree telnet wget net-tools  \
           tcpdump bash-completion sysstat chrony 

    echo -e "\033[32;32m 开始配置 k8s sysctl \033[0m \n" 
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
    modprobe br_netfilter && sysctl -p /etc/sysctl.d/k8s.conf
    systemctl enable chronyd && systemctl start chronyd && chronyc sources

    echo -e "\033[32;32m 开始配置系统ipvs \033[0m \n"

    cat > /etc/sysconfig/modules/ipvs.modules <<EOF
#!/bin/bash
modprobe -- ip_vs
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- ip_vs_sh
modprobe -- nf_conntrack
modprobe -- ip_tables
modprobe -- ip_set
modprobe -- xt_set
modprobe -- ipt_set
modprobe -- ipt_rpfilter
modprobe -- ipt_REJECT
modprobe -- ipip
EOF
    chmod 755 /etc/sysconfig/modules/ipvs.modules && bash /etc/sysconfig/modules/ipvs.modules && lsmod | grep -e ip_vs -e nf_conntrack
}

function Install_docker(){
    if [ -f /etc/docker/daemon.json ]; then
      echo -e "\033[32;32m 已完成docker安装 \033[0m \n" 
      return
    fi
    
    echo -e "\033[32;32m 开始安装docker \033[0m \n" 
    yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo 
    yum makecache fast
    yum install -y docker-ce-{{ .DockerVersion }} docker-ce-cli-{{ .DockerVersion }}

    echo -e "\033[32;32m 开始写 docker daemon.json\033[0m \n"
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<EOF 
{
  "exec-opts": [
    "native.cgroupdriver={{ default "systemd" .Cgroupdriver }}"
  ],
  "data-root": "/var/lib/docker",
  "ip-forward": true,
  "ip-masq": false,
  "iptables": false,
  "ipv6": false,
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
    systemctl enable docker && systemctl daemon-reload && systemctl restart docker
}

function Update_kernel(){
    uname -r | grep 5.7 &> /dev/null && echo -e "\033[32;32m 已完成内核升级 \033[0m \n" && return 

    echo -e "\033[32;32m 升级Centos7系统内核到5版本，解决Docker-ce版本兼容问题\033[0m \n"
    rpm --import https://www.elrepo.org/RPM-GPG-KEY-elrepo.org && \
    rpm -Uvh http://www.elrepo.org/elrepo-release-7.0-3.el7.elrepo.noarch.rpm 
    yum --disablerepo=\* --enablerepo=elrepo-kernel repolist && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml.x86_64 && \
    yum remove -y kernel-tools-libs.x86_64 kernel-tools.x86_64 && \
    yum --disablerepo=\* --enablerepo=elrepo-kernel install -y kernel-ml-tools.x86_64 && \
    grub2-set-default 0

#    wget https://cbs.centos.org/kojifiles/packages/kernel/4.9.221/37.el7/x86_64/kernel-4.9.221-37.el7.x86_64.rpm
#    rpm -ivh kernel-4.9.221-37.el7.x86_64.rpm
}

# 初始化顺序
echo -e "\033[32;32m 开始初始化结点 @{{ .HostIP }}@ \033[0m \n" 
Firewalld_process && \
Install_depend_environment && \
Install_docker && \
Update_kernel
`
)
