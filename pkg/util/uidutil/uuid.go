package uidutil

import (
	"github.com/google/uuid"
	"k8s.io/klog"
)

func GenerateId() string {
	u1, err := uuid.NewUUID()
	if err != nil {
		klog.Error("generate uuid error ", err)
		return err.Error()
	}
	return u1.String()
}
