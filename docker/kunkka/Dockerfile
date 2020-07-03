ARG CENTOS_BASE_VERSION=7.8
ARG K8S_BASE_VERSION=v1.18.5
ARG ETCD_BASE_VERSION=3.4.3-0

FROM k8s.gcr.io/kube-apiserver:${K8S_BASE_VERSION} as kube-apiserver-default
FROM registry.aliyuncs.com/google_containers/etcd:${ETCD_BASE_VERSION} as etcd-default

FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION}

COPY --from=etcd-default /usr/local/bin/etcd /usr/local/bin/
COPY --from=etcd-default /usr/local/bin/etcdctl /usr/local/bin/
COPY --from=kube-apiserver-default /usr/local/bin/kube-apiserver /usr/local/bin/

COPY bin/k9s /usr/local/bin/
COPY bin/kunkka-controller /usr/local/bin/