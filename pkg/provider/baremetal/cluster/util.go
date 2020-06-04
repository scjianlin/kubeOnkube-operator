package cluster

import (
	"fmt"
	"math"
	"net"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/ipallocator"
	"github.com/pkg/errors"
)

func GetNodeCIDRMaskSize(clusterCIDR string, maxNodePodNum int32) (int32, error) {
	if maxNodePodNum <= 0 {
		return 0, errors.New("maxNodePodNum must more than 0")
	}
	_, svcSubnetCIDR, err := net.ParseCIDR(clusterCIDR)
	if err != nil {
		return 0, errors.Wrap(err, "ParseCIDR error")
	}

	nodeCidrOccupy := math.Ceil(math.Log2(float64(maxNodePodNum)))
	nodeCIDRMaskSize := 32 - int(nodeCidrOccupy)
	ones, _ := svcSubnetCIDR.Mask.Size()
	if ones > nodeCIDRMaskSize {
		return 0, errors.New("clusterCIDR IP size is less than maxNodePodNum")
	}

	return int32(nodeCIDRMaskSize), nil
}

func GetServiceCIDRAndNodeCIDRMaskSize(clusterCIDR string, maxClusterServiceNum int32, maxNodePodNum int32) (string, int32, error) {
	if maxClusterServiceNum <= 0 || maxNodePodNum <= 0 {
		return "", 0, errors.New("maxClusterServiceNum or maxNodePodNum must more than 0")
	}
	_, svcSubnetCIDR, err := net.ParseCIDR(clusterCIDR)
	if err != nil {
		return "", 0, errors.Wrap(err, "ParseCIDR error")
	}

	size := ipallocator.RangeSize(svcSubnetCIDR)
	if int32(size) < maxClusterServiceNum {
		return "", 0, errors.New("clusterCIDR IP size is less than maxClusterServiceNum")
	}
	lastIP, err := ipallocator.GetIndexedIP(svcSubnetCIDR, int(size-1))
	if err != nil {
		return "", 0, errors.Wrap(err, "get last IP error")
	}

	maskSize := int(math.Ceil(math.Log2(float64(maxClusterServiceNum))))
	_, serviceCidr, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", lastIP.String(), 32-maskSize))

	nodeCidrOccupy := math.Ceil(math.Log2(float64(maxNodePodNum)))
	nodeCIDRMaskSize := 32 - int(nodeCidrOccupy)
	ones, _ := svcSubnetCIDR.Mask.Size()
	if ones > nodeCIDRMaskSize {
		return "", 0, errors.New("clusterCIDR IP size is less than maxNodePodNum")
	}

	return serviceCidr.String(), int32(nodeCIDRMaskSize), nil
}

func GetIndexedIP(subnet string, index int) (net.IP, error) {
	_, svcSubnetCIDR, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse service subnet CIDR %q", subnet)
	}

	dnsIP, err := ipallocator.GetIndexedIP(svcSubnetCIDR, index)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get %dth IP address from service subnet CIDR %s", index, svcSubnetCIDR.String())
	}

	return dnsIP, nil
}

// GetAPIServerCertSANs returns extra APIServer's certSANs need to pass kubeadm
func GetAPIServerCertSANs(c *devopsv1.Cluster) []string {
	certSANs := []string{
		"127.0.0.1",
		"localhost",
	}
	certSANs = append(certSANs, c.Spec.PublicAlternativeNames...)
	if c.Spec.Features.HA != nil {
		if c.Spec.Features.HA.TKEHA != nil {
			certSANs = append(certSANs, c.Spec.Features.HA.TKEHA.VIP)
		}
		if c.Spec.Features.HA.ThirdPartyHA != nil {
			certSANs = append(certSANs, c.Spec.Features.HA.ThirdPartyHA.VIP)
		}
	}
	for _, address := range c.Status.Addresses {
		certSANs = append(certSANs, address.Host)
	}

	return certSANs
}
