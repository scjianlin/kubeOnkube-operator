ARG CENTOS_BASE_VERSION=7.8
ARG K8S_BASE_VERSION=v1.18.5

FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION} as default

ENV K9S_SERVER_VERSION v0.21.2
RUN curl -fsSL https://github.com/derailed/k9s/releases/download/$K9S_SERVER_VERSION/k9s_Linux_x86_64.tar.gz -o k9s.tar.gz \
    && tar -C /usr/local -xzf k9s.tar.gz \
    && rm k9s.tar.gz

ENV CNI_PLUGINS_VERSION v0.8.6
RUN curl -fsSL https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz -o cni.tgz \
    && mkdir -p /opt/cni/bin/ \
    && tar -C /opt/cni/bin/ -xzf cni.tgz \
    && rm cni.tgz

FROM symcn.tencentcloudcr.com/symcn/kubernetes:${K8S_BASE_VERSION}

COPY --from=default /opt/cni/bin /opt/cni/bin
COPY --from=default /usr/local/k9s /usr/local/bin/
COPY bin/kunkka-controller /usr/local/bin/

COPY Dockerfile /Dockerfile