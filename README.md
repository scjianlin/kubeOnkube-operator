# Kunkka
Step lively now, your Admiral is on board!

# init tool

~~~shell
# 初始化二进制文件, 进入tools目录
$ cd tools
$ ./init.sh

# 进入项目根目录 
# 运行 fake apiserver
go run cmd/admin-controller/main.go fake -v 4 

# 运行正常后
$ ls -l k8s/cfg/fake-kubeconfig.yaml
-rw-------  1 xk  staff  276  6  3 10:23 k8s/cfg/fake-kubeconfig.yaml

# apply crd
$ export KUBECONFIG=k8s/cfg/fake-kubeconfig.yaml && kubectl apply -f config/crd/bases/ 
customresourcedefinition.apiextensions.k8s.io/clustercredentials.devops.gostship.io created
customresourcedefinition.apiextensions.k8s.io/clusters.devops.gostship.io created
customresourcedefinition.apiextensions.k8s.io/machines.devops.gostship.io created


# 运行 ctrl
go run cmd/admin-controller/main.go ctrl -v 4 --kubeconfig=k8s/cfg/fake-kubeconfig.yaml
~~~