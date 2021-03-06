package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/cidrutil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	"github.com/gostship/kunkka/pkg/util/uidutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"strconv"
)

var (
	ConfigMapName = "kunkka-api"
)

// Add ConfigMap data
func (m *Manager) AddRackCidr(c *gin.Context) {
	newRack := &model.Rack{}
	resp := responseutil.Gin{Ctx: c}

	// 获取创建Rack结构体
	r, err := resp.Bind(newRack)
	if err != nil {
		klog.Error("Http Bind ConfigMap error %v: ", err)
		resp.RespError("http Bind ConfigMap error")
		return
	}

	// 赋值UUID
	uid := uidutil.GenerateId()
	r.(*model.Rack).ID = uid

	// generate pod or host address
	rackNetAddr := r.(*model.Rack).RackCidr //10.28.0.0/22
	podList, hostList := cidrutil.GenerateCidr(rackNetAddr, r.(*model.Rack).RackCidrGw, r.(*model.Rack).PodNum, r.(*model.Rack).ServiceRoute, r.(*model.Rack).RackTag)
	r.(*model.Rack).HostAddr = hostList
	r.(*model.Rack).PodCidr = podList

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	listMap := []*model.Rack{}
	cm := &corev1.ConfigMap{}

	// 获取configMap
	err = cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: ConfigMapName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			metaCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigMapName,
					Namespace: ConfigMapName,
				},
				Data: map[string]string{"List": ""},
			}

			// 创建 comfilMap
			err := cli.Create(ctx, metaCm)
			if err != nil {
				klog.Errorf("failed to create rack configMaps, %s", err)
				resp.RespError("failed to create rack configMaps.")
				return
			}
			cm = metaCm
		}
	}

	klog.Info("re create ConfigMap Name: ", ConfigMapName)

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		resp.RespError("no configMap list!")
		return
	}
	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yaml to struct error!")
		return
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToRack, &listMap)
	if err != nil {
		klog.Errorf("Unmarshal json err", err)
		resp.RespError("Unmarshal list json error.")
		return
	}
	for _, rack := range listMap {
		if rack.RackCidr == r.(*model.Rack).RackCidr {
			// cidr already
			klog.Error("cidr %s is already:", r.(*model.Rack).RackCidr)
			resp.RespError(fmt.Sprintf("cidr %s is already", r.(*model.Rack).RackCidr))
			return
		}
	}
	// 将新数据添加到Map
	listMap = append(listMap, r.(*model.Rack))

	// 反解析json字符串
	strRackList, _ := json.MarshalIndent(listMap, "", "  ")

	// 写入configMap
	cm.Data["List"] = string(strRackList)

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update rack configMap.")
		resp.RespError("failed to update  rack configMap.")
		return
	}

	resp.RespSuccess(true, nil, "OK", 0)
}

// Get ConfigMap data
func (m *Manager) GetRackMap(c *gin.Context) {
	cidrName := c.DefaultQuery("rackCidr", "all")
	page := c.Query("page")
	limit := c.Query("limit")
	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	cms := []model.Rack{}

	cmList := &corev1.ConfigMap{}
	err := cli.Get(ctx, types.NamespacedName{
		Namespace: ConfigMapName,
		Name:      ConfigMapName,
	}, cmList)

	if err != nil {
		if apierrors.IsNotFound(err) {
			metaCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigMapName,
					Namespace: ConfigMapName,
				},
				Data: map[string]string{"List": ""},
			}

			// 创建 comfilMap
			err := cli.Create(ctx, metaCm)
			if err != nil {
				klog.Errorf("failed to create rack configMaps, %s", err)
				resp.RespError("failed to create rack configMaps.")
				return
			}
			cmList = metaCm
		}
	}

	data := cmList.Data["List"]

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error")
		return
	}

	rerr := json.Unmarshal(yamlToRack, &cms)
	if rerr != nil {
		klog.Errorf("failed to Unmarshal err: %v", rerr)
		resp.RespError("failed to Unmarshal error.")
		return
	}
	rackList := []model.Rack{}
	resultList := []model.Rack{}
	if cidrName == "all" {
		rackList = cms
	} else {
		for n := 0; n < len(cms); n++ {
			if cms[n].RackCidr == cidrName {
				rackList = append(rackList, cms[n])
			}
		}
	}
	// page list
	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)
	if len(rackList) > limitInt {
		if len(rackList) < (pageInt-1)*limitInt+limitInt {
			resultList = rackList[(pageInt-1)*limitInt:]
		} else {
			resultList = rackList[(pageInt-1)*limitInt : (pageInt-1)*limitInt+limitInt]
		}
	} else {
		resultList = rackList
	}
	resp.RespSuccess(true, nil, resultList, len(rackList))
}

//Update configMap data
func (m *Manager) UptConfigMap(c *gin.Context) {
	newRack := &model.Rack{}
	resp := responseutil.Gin{Ctx: c}
	// 获取创建Rack结构体
	r, err := resp.Bind(newRack)
	if err != nil {
		klog.Error("http Bind update ConfigMap error %v: ", err)
		resp.RespError("Update httpParams error.")
		return
	}

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	listMap := []*model.Rack{}
	cm := &corev1.ConfigMap{}

	// 获取configMap
	err = cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: ConfigMapName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Error("get ConfigMap %s error %v: ", ConfigMapName, err)
			resp.RespError("get configMap error.")
			return
		}
	}

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		resp.RespError("no ConfigMap list!")
		return
	}

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error.")
		return
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToRack, &listMap)
	if err != nil {
		klog.Errorf("Unmarshal json err", err)
		resp.RespError("Unmarshal json err")
		return
	}

	// 查找修改数据
	for i := 0; i < len(listMap); i++ {
		if listMap[i].ID == r.(*model.Rack).ID {
			listMap[i] = r.(*model.Rack) // update
		}
	}

	// 反解析json字符串
	makeList, _ := json.MarshalIndent(listMap, "", "  ")

	// 写入configMap
	cm.Data["List"] = string(makeList)

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update Rack configMap.")
		resp.RespError("failed to update Rack configMap.")
		return
	}

	resp.RespSuccess(true, nil, "OK", 0)
}

// Delete ConfigMap data
func (m *Manager) DelConfigMap(c *gin.Context) {
	newRack := &model.Rack{}
	resp := responseutil.Gin{Ctx: c}

	// 获取创建Rack结构体
	r, err := resp.Bind(newRack)
	if err != nil {
		klog.Error("bind delete ConfigMap error %v: ", err)
		resp.RespError("bind delete Params error")
		return
	}

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	listMap := []*model.Rack{}
	cm := &corev1.ConfigMap{}

	// 获取configMap
	err = cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: ConfigMapName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Error("get ConfigMap %s error %v: ", ConfigMapName, err)
			resp.RespError("get configMap error")
			return
		}
	}

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		resp.RespError("no ConfigMap list!")
		return
	}
	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error.")
		return
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToRack, &listMap)
	if err != nil {
		klog.Errorf("Unmarshal json err", err)
		resp.RespError("Unmarshal json err.")
		return
	}

	// 查找修改数据
	for i := 0; i < len(listMap); i++ {
		if listMap[i].ID == r.(*model.Rack).ID && listMap[i].RackCidr == r.(*model.Rack).RackCidr {
			listMap = append(listMap[:i], listMap[i+1:]...) //delete
		}
	}

	// 反解析json字符串
	strList, _ := json.MarshalIndent(listMap, "", "  ")

	// 写入configMap
	cm.Data["List"] = string(strList)

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update Rack configMap.")
		resp.RespError("failed to update Rack configMap.")
		return
	}

	resp.RespSuccess(true, nil, "OK", 0)
}

// get master rack
func (m *Manager) getMasterRack(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	cms := []model.Rack{}

	cmList := &corev1.ConfigMap{}
	err := cli.Get(ctx, types.NamespacedName{
		Namespace: ConfigMapName,
		Name:      ConfigMapName,
	}, cmList)

	if err != nil {
		klog.Error("Get ConfigMap error %v: ", err)
		resp.RespError("can't found rackcidr, please create!")
		return
	}

	data := cmList.Data["List"]

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error")
		return
	}

	rerr := json.Unmarshal(yamlToRack, &cms)
	if rerr != nil {
		klog.Errorf("failed to Unmarshal err: %v", rerr)
		resp.RespError("failed to Unmarshal error.")
		return
	}
	rackList := []model.Rack{}

	for _, rack := range cms {
		if rack.IsMaster == 1 {
			rackList = append(rackList, rack)
		}
	}
	resp.RespSuccess(true, "scuccess", rackList, len(rackList))
}

func (m *Manager) getHostRack(typeName string, c *gin.Context, clstype string) *model.Rack {
	resp := responseutil.Gin{Ctx: c}
	cli := m.Cluster.GetClient()
	ctx := context.Background()
	cms := []model.Rack{}

	cmList := &corev1.ConfigMap{}
	err := cli.Get(ctx, types.NamespacedName{
		Namespace: ConfigMapName,
		Name:      ConfigMapName,
	}, cmList)

	if err != nil {
		if apierrors.IsNotFound(err) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ConfigMapName,
				},
			}
			err := cli.Create(ctx, ns)
			if err != nil {
				klog.Errorf("get namespace:%s , error: %s", ConfigMapName, err)
				resp.RespError("get namespace error.")
			}
		}

		klog.Error("get configMap error %v: ", err)
		resp.RespError("can't found rackcidr, please create.")
	}

	data := cmList.Data["List"]

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error.")
	}

	rerr := json.Unmarshal(yamlToRack, &cms)
	if rerr != nil {
		klog.Errorf("failed to Unmarshal err: %v", rerr)
		resp.RespError("failed to Unmarshal err.")
	}

	result := &model.Rack{}
	for _, rack := range cms {
		if clstype == "Baremetal" {
			for _, hosts := range rack.HostAddr {
				if hosts.IPADDR == typeName {
					result = &rack
					return result
				}
			}
		} else { //全托管集群返回机柜信息
			if typeName == rack.RackTag {
				result = &rack
				return result
			}
		}
	}
	return result
}

// 更新机柜使用状态信息
func (m *Manager) UptRackStatePhase(rack string, machines string, podID string, state int) error {
	// 获取创建Rack结构体
	cli := m.Cluster.GetClient()
	ctx := context.Background()
	listMap := []*model.Rack{}
	cm := &corev1.ConfigMap{}

	// 获取configMap
	err := cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: ConfigMapName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Error("get ConfigMap %s error %v: ", ConfigMapName, err)
			return err
		}
	}

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		return errors.New("not found rack configMap Data!")
	}

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		return err
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToRack, &listMap)
	if err != nil {
		klog.Errorf("Unmarshal json err", err)
		return err
	}

	// 查找修改数据
	for i := 0; i < len(listMap); i++ {
		if listMap[i].RackTag == rack {
			for j := 0; i < len(listMap[i].HostAddr); j++ {
				if listMap[i].HostAddr[j].IPADDR == machines { //获取已经选择的machine节点
					listMap[i].HostAddr[j].UseState = state // 如果为0表示释放改CIDR资源;为1则正在使用该资源
				}
			}
			for k := 0; k < len(listMap[i].PodCidr); k++ {
				if listMap[i].PodCidr[k].ID == podID {
					listMap[i].PodCidr[k].UseState = state // 如果为0表示释放改CIDR资源;为1则正在使用该资源
				}
			}
		}
	}

	// 反解析json字符串
	makeList, _ := json.MarshalIndent(listMap, "", "  ")

	// 写入configMap
	cm.Data["List"] = string(makeList)

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update Rack configMap.")
		return err
	}
	return nil
}
