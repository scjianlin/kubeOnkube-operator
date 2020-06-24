ARG K8S_BASE_VERSION=v1.18.4
ARG CENTOS_BASE_VERSION=7.8

FROM k8s.gcr.io/kube-apiserver:${K8S_BASE_VERSION} as kube-apiserver-default
FROM k8s.gcr.io/kube-controller-manager:${K8S_BASE_VERSION} as kube-controller-manager-default
FROM k8s.gcr.io/kube-scheduler:${K8S_BASE_VERSION} as kube-scheduler-default
FROM k8s.gcr.io/kube-proxy:${K8S_BASE_VERSION} as kube-proxy-default


FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION}

COPY --from=kube-apiserver-default /usr/local/bin/kube-apiserver /usr/local/bin/
COPY --from=kube-controller-manager-default /usr/local/bin/kube-controller-manager /usr/local/bin/
COPY --from=kube-scheduler-default /usr/local/bin/kube-scheduler /usr/local/bin/
COPY --from=kube-proxy-default /usr/local/bin/kube-proxy /usr/local/bin/