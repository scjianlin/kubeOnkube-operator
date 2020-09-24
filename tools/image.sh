#!/usr/bin/env bash

set -xeuo pipefail

#v1.18.15
#K8s_Version=${1:-"v1.18.5"}
#Etcd_Version=${2:-"3.4.3-0"}
#CoreDns_Version=${3:-"1.6.7"}
#DstImagePrefix=${4:-"symcn.tencentcloudcr.com/symcn"}
#Pause_Version=${5:-"3.2"}

#v1.16.13
K8s_Version=${1:-"v1.16.14"}
Etcd_Version=${2:-"3.3.15-0"}
CoreDns_Version=${3:-"1.6.2"}
DstImagePrefix=${4:-"symcn.tencentcloudcr.com/symcn"}
Pause_Version=${5:-"3.1"}

docker pull k8s.gcr.io/kube-apiserver:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-apiserver:${K8s_Version} ${DstImagePrefix}/kube-apiserver:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-apiserver:${K8s_Version}

docker pull k8s.gcr.io/kube-controller-manager:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-controller-manager:${K8s_Version} ${DstImagePrefix}/kube-controller-manager:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-controller-manager:${K8s_Version}

docker pull k8s.gcr.io/kube-scheduler:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-scheduler:${K8s_Version} ${DstImagePrefix}/kube-scheduler:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-scheduler:${K8s_Version}

docker pull k8s.gcr.io/kube-proxy:${K8s_Version} && \
    docker tag k8s.gcr.io/kube-proxy:${K8s_Version} ${DstImagePrefix}/kube-proxy:${K8s_Version} && \
    docker push ${DstImagePrefix}/kube-proxy:${K8s_Version}

docker pull k8s.gcr.io/pause:${Pause_Version} && \
    docker tag k8s.gcr.io/pause:${Pause_Version} ${DstImagePrefix}/pause:${Pause_Version} && \
    docker push ${DstImagePrefix}/pause:${Pause_Version}

docker pull k8s.gcr.io/etcd:${Etcd_Version} && \
    docker tag k8s.gcr.io/etcd:${Etcd_Version} ${DstImagePrefix}/etcd:${Etcd_Version} && \
    docker push ${DstImagePrefix}/etcd:${Etcd_Version}

docker pull k8s.gcr.io/coredns:${CoreDns_Version} && \
    docker tag k8s.gcr.io/coredns:${CoreDns_Version} ${DstImagePrefix}/coredns:${CoreDns_Version} && \
    docker push ${DstImagePrefix}/coredns:${CoreDns_Version}

#k8s.gcr.io/kube-controller-manager:v1.18.5
#k8s.gcr.io/kube-scheduler:v1.18.5
#k8s.gcr.io/kube-proxy:v1.18.5
#k8s.gcr.io/pause:3.2
#k8s.gcr.io/etcd:3.4.3-0
#k8s.gcr.io/coredns:1.6.7