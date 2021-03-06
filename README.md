# Kunkka

Kunkka 是一个自动化部署高可用kubernetes的operator

# 特性

- 云原生架构，crd+controller，采用声明式api描述一个集群的最终状态
- 支持裸金属和master组件托管两种方式部署集群
- 可以启用一个fake-cluster，解决裸金属第一次部署集群没有元集群问题
- 无坑版100年集群证书，kubelet自动生成证书
- 除kubelet外集群组件全部容器化部署，componentstatuses可以发现三个etcd
- 支持coredns, flannel，metrics-server等 addons 模板化部署

# 安装部署

## 准备

下载fake-cluster需要二进制文件，启动fake-cluster

```bash
# 下载二进制文件, 进入tools目录
$ cd tools
$ ./init.sh

# 进入项目根目录  运行 fake apiserver
$ cd ..
$ go run cmd/admin-controller/main.go fake --baseBinDir k8s/bin --rootDir k8s -v 4 
$ export KUBECONFIG=k8s/cfg/fake-kubeconfig.yaml

# 运行正常后
$ cat k8s/cfg/fake-kubeconfig.yaml
apiVersion: v1
clusters:
- cluster:
    server: 127.0.0.1:18080
  name: fake-cluster
contexts:
- context:
    cluster: fake-cluster
    user: devops
  name: devops@fake-cluster
current-context: devops@fake-cluster
kind: Config
preferences: {}
users:
- name: devops
  user: {}
```

## 运行

本地运行
```bash
# apply crd
$ export KUBECONFIG=k8s/cfg/fake-kubeconfig.yaml && kubectl apply -f manifests/crds/
customresourcedefinition.apiextensions.k8s.io/clustercredentials.devops.gostship.io created
customresourcedefinition.apiextensions.k8s.io/clusters.devops.gostship.io created
customresourcedefinition.apiextensions.k8s.io/machines.devops.gostship.io created

# 运行
$ go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig=k8s/cfg/fake-kubeconfig.yaml
```
docker 运行
```bash
docker run --name fake-cluster -d --restart=always \
   --net="host" \
   --pid="host" \
   -v /root/kunkka/k8s:/kunkka \
   symcn.tencentcloudcr.com/symcn/kunkka:v0.0.3-dev11 \
   kunkka-controller fake --rootDir /kunkka -v 4

docker run --name kunkka-controller -d --restart=always \
   --net="host" \
   --pid="host" \
   -v /root/kunkka/k8s:/kunkka \
   symcn.tencentcloudcr.com/symcn/kunkka:v0.0.3-dev11 \
   kunkka-controller ctrl -v 4 --kubeconfig=/kunkka/cfg/fake-kubeconfig.yaml

export KUBECONFIG=/root/kunkka/k8s/cfg/fake-kubeconfig.yaml
```
指定裸金属运行，部署托管集群
```shell
# 进入安装结点目录
cd /root/kunkka/k8s/cfg
# 导入裸金属集群kubeconfig
ls -l /root/kunkka/k8s/cfg/bt-kubeconfig.yaml

# 运行
docker stop kunkka-controller && docker rm kunkka-controller

docker run --name meta-controller -d --restart=always \
   --net="host" \
   --pid="host" \
   -v /root/kunkka/k8s:/kunkka \
   -v /etc/hosts:/etc/hosts \
   symcn.tencentcloudcr.com/symcn/kunkka:v0.0.3-dev11 \
   kunkka-controller ctrl -v 2 --kubeconfig=/kunkka/cfg/bt-kubeconfig.yaml
```

#### Kunkka API 运行
```bash
# API的运行依赖Meta ApiServer!
export KUBECONFIG=manifests/fake/fake.yaml  
$ go run  cmd/admin-api/main.go api
```


#### 容器部署
charts目录中kunkka-api 为kunkka的API服务  
charts目录中kunkka-console 为kunkka的Console控制台  
charts目录中kunkka-controller 为kunkka的Controller  
charts目录中prometheus-operator 为容器集群的监控组件

#### 安装说明:
1. 首先需要有集群(可以使用fake创建)  
2. 有了集群之后helm instll kunkka-api/kunkka-controller/prometheus-operator 安装对应的程序即可


# 计划

- [x]  打通元集群及托管集群service网络，以支持聚合apiserver
- [x]  支持 helm v3 部署 addons