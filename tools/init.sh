#!/bin/bash

set -e  # exit immediately on error
set -x  # display all commands

PACKAGE_DARWIN="../k8sbin/darwin"
PACKAGE_LINUX="../k8sbin/linux"

if [ ! -d ${PACKAGE_DARWIN} ]; then
	if [ ! -f kubebuilder_2.3.1_darwin_amd64.tar.gz ]; then
		wget https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.1/kubebuilder_2.3.1_darwin_amd64.tar.gz
	fi

	tar -xf kubebuilder_2.3.1_darwin_amd64.tar.gz
	mkdir -p ${PACKAGE_DARWIN}
    cp kubebuilder_2.3.1_darwin_amd64/bin/* ${PACKAGE_DARWIN}
fi

if [ ! -d ${PACKAGE_LINUX} ]; then
	if [ ! -f kubebuilder_2.3.1_linux_amd64.tar.gz ]; then
		wget https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.1/kubebuilder_2.3.1_linux_amd64.tar.gz
	fi

	tar -xf kubebuilder_2.3.1_linux_amd64.tar.gz
	mkdir -p ${PACKAGE_LINUX}
    cp kubebuilder_2.3.1_linux_amd64/bin/* ${PACKAGE_LINUX}
fi


echo "all done."