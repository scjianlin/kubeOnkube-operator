package v1

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gostship/kunkka/pkg/util/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (in *Cluster) SetCondition(newCondition ClusterCondition) {
	var conditions []ClusterCondition

	exist := false

	if newCondition.LastProbeTime.IsZero() {
		newCondition.LastProbeTime = metav1.Now()
	}
	for _, condition := range in.Status.Conditions {
		if condition.Type == newCondition.Type {
			exist = true
			if newCondition.LastTransitionTime.IsZero() {
				newCondition.LastTransitionTime = condition.LastTransitionTime
			}
			condition = newCondition
		}
		conditions = append(conditions, condition)
	}

	if !exist {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		conditions = append(conditions, newCondition)
	}

	in.Status.Conditions = conditions
}

func (in *ClusterMachine) SSH() (*ssh.SSH, error) {
	sshConfig := &ssh.Config{
		User:        in.Username,
		Host:        in.IP,
		Port:        int(in.Port),
		Password:    in.Password,
		PrivateKey:  in.PrivateKey,
		PassPhrase:  in.PassPhrase,
		DialTimeOut: time.Second,
		Retry:       0,
	}
	return ssh.New(sshConfig)
}

func (in *Cluster) Address(addrType AddressType) *ClusterAddress {
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			return &one
		}
	}

	return nil
}

func (in *Cluster) AddAddress(addrType AddressType, host string, port int32) {
	addr := ClusterAddress{
		Type: addrType,
		Host: host,
		Port: port,
	}
	for _, one := range in.Status.Addresses {
		if one == addr {
			return
		}
	}
	in.Status.Addresses = append(in.Status.Addresses, addr)
}

func (in *Cluster) RemoveAddress(addrType AddressType) {
	var addrs []ClusterAddress
	for _, one := range in.Status.Addresses {
		if one.Type == addrType {
			continue
		}
		addrs = append(addrs, one)
	}
	in.Status.Addresses = addrs
}

func (in *Cluster) Host() (string, error) {
	addrs := make(map[AddressType][]ClusterAddress)
	for _, one := range in.Status.Addresses {
		addrs[one.Type] = append(addrs[one.Type], one)
	}

	var address *ClusterAddress
	if len(addrs[AddressInternal]) != 0 {
		address = &addrs[AddressInternal][rand.Intn(len(addrs[AddressInternal]))]
	} else if len(addrs[AddressAdvertise]) != 0 {
		address = &addrs[AddressAdvertise][rand.Intn(len(addrs[AddressAdvertise]))]
	} else {
		if len(addrs[AddressReal]) != 0 {
			address = &addrs[AddressReal][rand.Intn(len(addrs[AddressReal]))]
		}
	}

	if address == nil {
		return "", errors.New("can't find valid address")
	}

	return fmt.Sprintf("%s:%d", address.Host, address.Port), nil
}

func (in *Machine) SetCondition(newCondition MachineCondition) {
	var conditions []MachineCondition

	exist := false

	if newCondition.LastProbeTime.IsZero() {
		newCondition.LastProbeTime = metav1.Now()
	}
	for _, condition := range in.Status.Conditions {
		if condition.Type == newCondition.Type {
			exist = true
			if newCondition.LastTransitionTime.IsZero() {
				newCondition.LastTransitionTime = condition.LastTransitionTime
			}
			condition = newCondition
		}
		conditions = append(conditions, condition)
	}

	if !exist {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		conditions = append(conditions, newCondition)
	}

	in.Status.Conditions = conditions
}

func (in *MachineSpec) SSH() (*ssh.SSH, error) {
	sshConfig := &ssh.Config{
		User:        in.Username,
		Host:        in.IP,
		Port:        int(in.Port),
		Password:    in.Password,
		PrivateKey:  in.PrivateKey,
		PassPhrase:  in.PassPhrase,
		DialTimeOut: time.Second,
		Retry:       0,
	}
	return ssh.New(sshConfig)
}
