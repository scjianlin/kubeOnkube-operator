package cidrutil

import (
	"fmt"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	v1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/uidutil"
	"net"
	"strings"
)

//var (
//	Route = "10.27.248.0/22" //改成动态输入
//)

// generate subnet
func generate(cidr string) (*[]string, *[]string, string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, "", err
	}

	var ipList []string
	var hostList []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		hostList = append(hostList, ip.String())
		ipStr := strings.Split(ip.String(), ".")
		if ipStr[3] == "0" {
			// get subnet cidr
			ipList = append(ipList, ip.String())
		}
	}
	mask := genMaskString(ipnet.Mask)
	return &ipList, &hostList, mask, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// get pod list of subnet
func getPodCidr(ip string, podNum int, svcRoute string) []*model.PodAddr {
	ips := strings.Split(ip, ".")
	podCidr := []*model.PodAddr{}
	for i := 0; i < int(255/podNum); i++ {
		pod := &model.PodAddr{}
		pod.DefaultRoute = svcRoute
		pod.RangeStart = fmt.Sprintf("%s.%s.%s.%d", ips[0], ips[1], ips[2], i*podNum+1)
		pod.RangeEnd = fmt.Sprintf("%s.%s.%s.%d", ips[0], ips[1], ips[2], +i*podNum+podNum)
		pod.ID = uidutil.GenerateId()
		podCidr = append(podCidr, pod)
	}
	return podCidr
}

func genMaskString(m []byte) string {
	if len(m) != 4 {
		panic("ipv4Mask: len must be 4 bytes")
	}
	return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
}

func GenerateCidr(cidr string, gw string, podNum int, svcRoute string, rack string) ([]*v1.ClusterCni, []*model.HostAddr) {
	pod, host, mask, _ := generate(cidr)
	rackpodList := []*model.PodAddr{}
	for _, ip := range *pod {
		res := getPodCidr(ip, podNum, svcRoute)
		for _, cidr := range res {
			rackpodList = append(rackpodList, cidr)
		}
	}
	hostList := *host
	hostNum := len(rackpodList)
	rackHost := hostList[len(hostList)-3-hostNum : len(hostList)-3]
	//hostGw := hostList[len(hostList)-2]

	podlist := []*v1.ClusterCni{}
	hostlist := []*model.HostAddr{}

	for _, v := range rackpodList {
		p := &v1.ClusterCni{
			ID:           uidutil.GenerateId(),
			Subnet:       cidr,
			RangeStart:   v.RangeStart,
			RangeEnd:     v.RangeEnd,
			DefaultRoute: svcRoute,
			RackTag:      rack,
			UseState:     0,
			GW:           gw,
		}
		podlist = append(podlist, p)
	}
	for _, v := range rackHost {
		h := &model.HostAddr{
			ID:       uidutil.GenerateId(),
			IPADDR:   v,
			NetMask:  mask,
			GateWay:  gw,
			UseState: 0,
			//DnsServer: []string{"10.27.0.2", "10.27.0.202"},
		}
		hostlist = append(hostlist, h)
	}
	return podlist, hostlist
}
