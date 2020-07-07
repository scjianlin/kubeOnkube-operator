FROM centos:7.8.2003

RUN yum update -y && yum install -y \
    bridge-utils \
    iptables \
    lsof \
    curl \
    net-tools \
    tree \
    tcpdump \
    vim \
    netsgtat \
    ipset \
    tcptraceroute \
    && yum clean all \
    && rm -rf /var/cache/yum/*