package machine

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/thoas/go-funk"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
)

const (
	ReasonFailedProcess     = "FailedProcess"
	ReasonWaitingProcess    = "WaitingProcess"
	ReasonSuccessfulProcess = "SuccessfulProcess"
	ReasonSkipProcess       = "SkipProcess"

	ConditionTypeDone = "EnsureDone"
)

// Provider defines a set of response interfaces for specific machine
// types in machine management.
type Provider interface {
	Name() string

	Validate(machine *devopsv1.Machine) field.ErrorList

	PreCreate(machine *devopsv1.Machine) error
	AfterCreate(machine *devopsv1.Machine) error

	OnCreate(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error
	OnUpdate(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error
	OnDelete(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error
}

var _ Provider = &DelegateProvider{}

type Handler func(context.Context, *devopsv1.Machine, *provider.Cluster) error

type DelegateProvider struct {
	ProviderName string

	ValidateFunc    func(machine *devopsv1.Machine) field.ErrorList
	PreCreateFunc   func(machine *devopsv1.Machine) error
	AfterCreateFunc func(machine *devopsv1.Machine) error

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

func (p *DelegateProvider) Validate(machine *devopsv1.Machine) field.ErrorList {
	if p.ValidateFunc != nil {
		return p.ValidateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) PreCreate(machine *devopsv1.Machine) error {
	if p.PreCreateFunc != nil {
		return p.PreCreateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) AfterCreate(machine *devopsv1.Machine) error {
	if p.AfterCreateFunc != nil {
		return p.AfterCreateFunc(machine)
	}

	return nil
}

func (p *DelegateProvider) OnCreate(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error {
	condition, err := p.getCreateCurrentCondition(machine)
	if err != nil {
		return err
	}

	now := metav1.Now()
	if cluster.Spec.Features.SkipConditions != nil &&
		funk.ContainsString(cluster.Spec.Features.SkipConditions, condition.Type) {
		machine.SetCondition(devopsv1.MachineCondition{
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
		klog.Infof("OnCreate", "handler", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
			"machineName", machine.Name)
		err = f(ctx, machine, cluster)
		if err != nil {
			machine.SetCondition(devopsv1.MachineCondition{
				Type:          condition.Type,
				Status:        devopsv1.ConditionFalse,
				LastProbeTime: now,
				Message:       err.Error(),
				Reason:        ReasonFailedProcess,
			})
			machine.Status.Reason = ReasonFailedProcess
			machine.Status.Message = err.Error()
			return nil
		}

		machine.SetCondition(devopsv1.MachineCondition{
			Type:               condition.Type,
			Status:             devopsv1.ConditionTrue,
			LastProbeTime:      now,
			LastTransitionTime: now,
			Reason:             ReasonSuccessfulProcess,
		})
	}

	nextConditionType := p.getNextConditionType(condition.Type)
	if nextConditionType == ConditionTypeDone {
		machine.Status.Phase = devopsv1.MachineRunning
	} else {
		machine.SetCondition(devopsv1.MachineCondition{
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

func (p *DelegateProvider) OnUpdate(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error {
	for _, f := range p.UpdateHandlers {
		klog.Infof("OnUpdate", "handler", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
			"machineName", machine.Name)
		err := f(ctx, machine, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DelegateProvider) OnDelete(ctx context.Context, machine *devopsv1.Machine, cluster *provider.Cluster) error {
	for _, f := range p.DeleteHandlers {
		klog.Infof("OnDelete", "handler", runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
			"machineName", machine.Name)
		err := f(ctx, machine, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h Handler) name() string {
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

	return next.name()
}

func (p *DelegateProvider) getCreateHandler(conditionType string) Handler {
	for _, f := range p.CreateHandlers {
		if conditionType == f.name() {
			return f
		}
	}

	return nil
}

func (p *DelegateProvider) getCreateCurrentCondition(c *devopsv1.Machine) (*devopsv1.MachineCondition, error) {
	if c.Status.Phase == devopsv1.MachineRunning {
		return nil, errors.New("machine phase is running now")
	}
	if len(p.CreateHandlers) == 0 {
		return nil, errors.New("no create handlers")
	}

	if len(c.Status.Conditions) == 0 {
		return &devopsv1.MachineCondition{
			Type:          p.CreateHandlers[0].name(),
			Status:        devopsv1.ConditionUnknown,
			LastProbeTime: metav1.Now(),
			Message:       "waiting process",
			Reason:        ReasonWaitingProcess,
		}, nil
	}

	for _, condition := range c.Status.Conditions {
		if condition.Status == devopsv1.ConditionFalse || condition.Status == devopsv1.ConditionUnknown {
			return &condition, nil
		}
	}

	return nil, errors.New("no condition need process")
}
