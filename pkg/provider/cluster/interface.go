package cluster

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/provider"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/klog"
)

const (
	ReasonFailedProcess     = "FailedProcess"
	ReasonWaitingProcess    = "WaitingProcess"
	ReasonSuccessfulProcess = "SuccessfulProcess"
	ReasonSkipProcess       = "SkipProcess"

	ConditionTypeDone = "EnsureDone"
)

// Provider defines a set of response interfaces for specific cluster
// types in cluster management.
type Provider interface {
	Name() string

	RegisterHandler(mux *mux.PathRecorderMux)

	Validate(cluster *provider.Cluster) field.ErrorList

	PreCreate(cluster *provider.Cluster) error
	AfterCreate(cluster *provider.Cluster) error

	OnCreate(ctx context.Context, cluster *provider.Cluster) error
	OnUpdate(ctx context.Context, cluster *provider.Cluster) error
	OnDelete(ctx context.Context, cluster *provider.Cluster) error
}

var _ Provider = &DelegateProvider{}

type Handler func(context.Context, *provider.Cluster) error

type DelegateProvider struct {
	ProviderName string

	ValidateFunc    func(cluster *provider.Cluster) field.ErrorList
	PreCreateFunc   func(cluster *provider.Cluster) error
	AfterCreateFunc func(cluster *provider.Cluster) error

	CreateHandlers []Handler
	DeleteHandlers []Handler
	UpdateHandlers []Handler
}

func (p *DelegateProvider) Name() string {
	if p.ProviderName == "" {
		return "unknown"
	}
	return p.ProviderName
}

func (p *DelegateProvider) RegisterHandler(mux *mux.PathRecorderMux) {
}

func (p *DelegateProvider) Validate(cluster *provider.Cluster) field.ErrorList {
	if p.ValidateFunc != nil {
		return p.ValidateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) PreCreate(cluster *provider.Cluster) error {
	if p.PreCreateFunc != nil {
		return p.PreCreateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) AfterCreate(cluster *provider.Cluster) error {
	if p.AfterCreateFunc != nil {
		return p.AfterCreateFunc(cluster)
	}

	return nil
}

func (p *DelegateProvider) OnCreate(ctx context.Context, cluster *provider.Cluster) error {
	condition, err := p.getCreateCurrentCondition(cluster)
	if err != nil {
		return err
	}

	now := metav1.Now()
	if cluster.Spec.Features.SkipConditions != nil &&
		funk.ContainsString(cluster.Spec.Features.SkipConditions, condition.Type) {
		cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSkipProcess,
		})
	} else {
		f := p.getCreateHandler(condition.Type)
		if f == nil {
			return fmt.Errorf("can't get handler by %s", condition.Type)
		}
		handlerName := f.Name()
		klog.Infof("clusterName: %s OnCreate handler: %s", cluster.Name, handlerName)
		err = f(ctx, cluster)
		if err != nil {
			klog.Errorf("cluster: %s OnCreate handler: %s err: %+v", cluster.Name, handlerName, err)
			cluster.SetCondition(devopsv1.ClusterCondition{
				Type:          condition.Type,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			cluster.Status.Reason = ReasonFailedProcess
			cluster.Status.Message = err.Error()
			return nil
		}

		cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	nextConditionType := p.getNextConditionType(condition.Type)
	if nextConditionType == ConditionTypeDone {
		cluster.Status.Phase = devopsv1.ClusterRunning
	} else {
		cluster.SetCondition(devopsv1.ClusterCondition{
			Type:               nextConditionType,
			Status:             devopsv1.ConditionUnknown,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Message:            "waiting process",
			Reason:             ReasonWaitingProcess,
		})
	}

	return nil
}

func (p *DelegateProvider) OnUpdate(ctx context.Context, cluster *provider.Cluster) error {
	for _, f := range p.UpdateHandlers {
		klog.Infof("clusterName: %s OnUpdate handler: %s", cluster.Name,
			runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		err := f(ctx, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DelegateProvider) OnDelete(ctx context.Context, cluster *provider.Cluster) error {
	for _, f := range p.DeleteHandlers {
		klog.Infof("clusterName: %s OnDelete handler: %s", cluster.Name,
			runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		err := f(ctx, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h Handler) Name() string {
	name := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	i := strings.Index(name, "Ensure")
	if i == -1 {
		return ""
	}
	return strings.TrimSuffix(name[i:], "-fm")
}

func (p *DelegateProvider) getNextConditionType(conditionType string) string {
	var (
		i int
		f Handler
	)
	for i, f = range p.CreateHandlers {
		name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		if strings.Contains(name, conditionType) {
			break
		}
	}
	if i == len(p.CreateHandlers)-1 {
		return ConditionTypeDone
	}
	next := p.CreateHandlers[i+1]

	return next.Name()
}

func (p *DelegateProvider) getCreateHandler(conditionType string) Handler {
	for _, f := range p.CreateHandlers {
		if conditionType == f.Name() {
			return f
		}
	}

	return nil
}

func (p *DelegateProvider) getCreateCurrentCondition(c *provider.Cluster) (*devopsv1.ClusterCondition, error) {
	// if c.Status.Phase == devopsv1.ClusterRunning {
	// 	return nil, errors.New("cluster phase is running now")
	// }

	if len(p.CreateHandlers) == 0 {
		return nil, errors.New("no create handlers")
	}

	if len(c.Status.Conditions) == 0 {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[0].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	for _, condition := range c.Status.Conditions {
		// if condition.Type == "EnsureKubeadmInitWaitControlPlanePhase" {
		// 	return &condition, nil
		// }
		if condition.Status == devopsv1.ConditionFalse || condition.Status == devopsv1.ConditionUnknown {
			return &condition, nil
		}
	}

	if len(c.Status.Conditions) < len(p.CreateHandlers) {
		return &devopsv1.ClusterCondition{
			Type:          p.CreateHandlers[len(c.Status.Conditions)].Name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	return nil, errors.New("no condition need process")
}
