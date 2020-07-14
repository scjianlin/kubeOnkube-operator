ARG CENTOS_BASE_VERSION=v7.8.1
ARG K8S_BASE_VERSION=v1.18.5

FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION} as default

ENV K9S_SERVER_VERSION v0.21.2
RUN curl -fsSL https://github.com/derailed/k9s/releases/download/$K9S_SERVER_VERSION/k9s_Linux_x86_64.tar.gz -o k9s.tar.gz \
    && tar -C /usr/local/bin -xzf k9s.tar.gz \
    && rm k9s.tar.gz

ENV CNI_PLUGINS_VERSION v0.8.6
RUN curl -fsSL https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGINS_VERSION/cni-plugins-linux-amd64-$CNI_PLUGINS_VERSION.tgz -o cni.tgz \
    && mv cni.tgz /usr/local/bin/

ENV HELM_VERSION v3.2.4
RUN curl -fsSL https://get.helm.sh/helm-$HELM_VERSION-linux-amd64.tar.gz -o helm.tar.gz \
    && tar -C /tmp/ -xzf helm.tar.gz \
    && mv /tmp/linux-amd64/helm /usr/local/bin/ \
    && rm helm.tar.gz

FROM symcn.tencentcloudcr.com/symcn/kubernetes:${K8S_BASE_VERSION}

COPY --from=default /usr/local/bin/cni.tgz /usr/local/bin/
COPY --from=default /usr/local/bin/k9s /usr/local/bin/
COPY --from=default /usr/local/bin/helm /usr/local/bin/
COPY bin/kunkka-controller /usr/local/bin/
