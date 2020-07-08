ARG CENTOS_BASE_VERSION=v7.8.1
ARG METRICS_SERVER_BASE_VERSION=v0.3.6

FROM k8s.gcr.io/metrics-server:${METRICS_SERVER_BASE_VERSION} as default

FROM symcn.tencentcloudcr.com/symcn/centos-base:${CENTOS_BASE_VERSION}

COPY --from=default /metrics-server /

USER 65534

ENTRYPOINT ["/metrics-server"]
