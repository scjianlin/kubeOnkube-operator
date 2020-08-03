package apimanager

import (
	"context"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"strconv"
)

// Get ConfigMap data
func (m *APIManager) GetPodCidr(c *gin.Context) {
	cidrName := c.DefaultQuery("rackCidr", "all")
	resp := responseutil.Gin{Ctx: c}
	page := c.Query("page")
	limit := c.Query("limit")

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
				klog.Errorf("create namespace:%s , error: %s", ConfigMapName, err)
				resp.RespError("create namespace error.")
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
	podList := []*model.PodAddrList{}
	resultList := []*model.PodAddrList{}

	if cidrName == "all" {
		for _, rack := range cms {
			for _, pod := range rack.PodCidr {
				podmsg := &model.PodAddrList{
					ID:           pod.ID,
					RangeStart:   pod.RangeStart,
					RangeEnd:     pod.RangeEnd,
					DefaultRoute: pod.DefaultRoute,
					RackTag:      rack.RackTag,
					RackCidr:     rack.RackCidr,
				}
				podList = append(podList, podmsg)
			}
		}
	} else {
		for n := 0; n < len(cms); n++ {
			if cms[n].RackCidr == cidrName {
				for _, pod := range cms[n].PodCidr {
					podmsg := &model.PodAddrList{
						ID:           pod.ID,
						RangeStart:   pod.RangeStart,
						RangeEnd:     pod.RangeEnd,
						DefaultRoute: pod.DefaultRoute,
						RackTag:      cms[n].RackTag,
						RackCidr:     cms[n].RackCidr,
					}
					podList = append(podList, podmsg)
				}
			}
		}
	}

	// page list
	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)
	if len(podList) > limitInt {
		if len(podList) < (pageInt-1)*limitInt+limitInt {
			resultList = podList[(pageInt-1)*limitInt:]
		} else {
			resultList = podList[(pageInt-1)*limitInt : (pageInt-1)*limitInt+limitInt]
		}
	} else {
		resultList = podList
	}
	resp.RespSuccess(true, nil, resultList, len(podList))
}
