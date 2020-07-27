package apimanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"net/http"
)

var (
	ConfigMapName = "kunkka-api"
)

// Add ConfigMap data
func (m *APIManager) AddRackCidr(c *gin.Context) {
	NRack := &model.Rack{}

	// 获取创建Rack结构体
	r, err := Bind(NRack, c)
	if err != nil {
		klog.Error("Http Bind ConfigMap error %v: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Http Bind ConfigMap error",
			"data":    nil,
		})
		return
	}

	// 赋值UUID
	uid, err := generateId()
	if err != nil {
		klog.Errorf("", err)
	}
	r.(*model.Rack).ID = uid

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	listMap := []*model.Rack{}
	cm := &corev1.ConfigMap{}

	// 获取configMap
	err = cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: ConfigMapName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			MetaCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigMapName,
					Namespace: ConfigMapName,
				},
				Data: map[string]string{"List": ""},
			}

			// 创建 comfilMap
			err := cli.Create(ctx, MetaCm)
			if err != nil {
				klog.Errorf("failed to create rack conmaps, %s", err)
				return
			}
			cm = MetaCm
		}
	}

	klog.Info("re create ConfigMap Name: ", ConfigMapName)

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		return
	}
	// 将yaml转换为json
	yamlRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		fmt.Println("yamlToJson error", err)
	}
	// 转换为结构体
	err = json.Unmarshal(yamlRack, &listMap)
	if err != nil {
		fmt.Println("json err", err)
	}

	// 将新数据添加到Map
	listMap = append(listMap, r.(*model.Rack))

	// 反解析json字符串
	makeList, _ := json.MarshalIndent(listMap, "", "  ")

	// 写入configMap
	cm.Data["List"] = string(makeList)

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update Rack configMap.")
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"success": true,
		"message": nil,
		"data":    "OK",
	})
}

// Get ConfigMap data
func (m *APIManager) GetRackMap(c *gin.Context) {
	cidrName := c.DefaultQuery("rack_cidr", "all")

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
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"success":   false,
			"message":   "Can't found RackCidr by this IP",
			"resultMap": nil,
		})
		return
	}

	data := cmList.Data["List"]

	// 将yaml转换为json
	yamlRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		fmt.Println("yamlToJson error", err)
	}

	rerr := json.Unmarshal(yamlRack, &cms)
	if rerr != nil {
		klog.Errorf("failed to Unmarshal err: %v", rerr)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"message":   rerr.Error(),
			"resultMap": nil,
		})
	}
	resuleList := []model.Rack{}
	if cidrName == "all" {
		resuleList = cms
	} else {
		for n := 0; n < len(cms); n++ {
			if cms[n].RackCidr == cidrName {
				resuleList = append(resuleList, cms[n])
			}
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     nil,
		"items":       resuleList,
		"total_count": len(resuleList),
	})

}

//Update configMap data
func (m *APIManager) UptConfigMap(c *gin.Context) {
	NewRack := &model.Rack{}

	// 获取创建Rack结构体
	r, err := Bind(NewRack, c)
	if err != nil {
		klog.Error("Http Bind ConfigMap error %v: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Http Bind ConfigMap error",
			"data":    nil,
		})
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "get ConfigMap error",
				"data":    nil,
			})
			return
		}
	}

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		return
	}
	// 将yaml转换为json
	yamlRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		fmt.Println("yamlToJson error", err)
	}
	// 转换为结构体
	err = json.Unmarshal(yamlRack, &listMap)
	if err != nil {
		fmt.Println("json err", err)
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
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"success": true,
		"message": nil,
		"data":    "OK",
	})
}

// Delete ConfigMap data
func (m *APIManager) DelConfigMap(c *gin.Context) {
	NewRack := &model.Rack{}

	// 获取创建Rack结构体
	r, err := Bind(NewRack, c)
	if err != nil {
		klog.Error("Http Bind ConfigMap error %v: ", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Http Bind ConfigMap error",
			"data":    nil,
		})
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "get ConfigMap error",
				"data":    nil,
			})
			return
		}
	}

	// 获取confiMap的数据
	data, ok := cm.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		return
	}
	// 将yaml转换为json
	yamlRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		fmt.Println("yamlToJson error", err)
	}
	// 转换为结构体
	err = json.Unmarshal(yamlRack, &listMap)
	if err != nil {
		fmt.Println("json err", err)
	}

	// 查找修改数据
	for i := 0; i < len(listMap); i++ {
		if listMap[i].ID == r.(*model.Rack).ID && listMap[i].RackCidr == r.(*model.Rack).RackCidr {
			listMap = append(listMap[:i], listMap[i+1:]...) //delete
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
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"success": true,
		"message": nil,
		"data":    "OK",
	})
}
