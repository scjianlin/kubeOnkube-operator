ARG CENTOS_BASE_VERSION=v7.8.2
ARG ETCD_BASE_VERSION=v3.3.15
FROM quay.io/coreos/etcd:${ETCD_BASE_VERSION} as etcd-default
FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION} as default
ENV K8S_SERVER_VERSION v1.16.14
RUN export https_proxy=http://10.12.217.238:7890 http_proxy=http://10.12.217.238:7890 all_proxy=socks5://10.12.217.238:7890
RUN curl -fsSL https://dl.k8s.io/$K8S_SERVER_VERSION/kubernetes-server-linux-amd64.tar.gz  -o kubernetes-server-linux-amd64.tar.gz \
    && tar -C /usr/local -xzf kubernetes-server-linux-amd64.tar.gz \
    && rm kubernetes-server-linux-amd64.tar.gz

FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION}

COPY --from=default /usr/local/kubernetes/server/bin/kube-apiserver \
                    /usr/local/kubernetes/server/bin/kubeadm  \
                    /usr/local/kubernetes/server/bin/kubectl \
                    /usr/local/kubernetes/server/bin/kubelet \
                    /usr/local/kubernetes/server/bin/kube-scheduler \
                    /usr/local/kubernetes/server/bin/kube-controller-manager \
                    /usr/local/bin/

COPY --from=etcd-default /usr/local/bin/etcd \
                          /usr/local/bin/etcdctl \
                          /usr/local/bin/
