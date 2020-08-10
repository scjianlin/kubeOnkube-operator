/*
Copyright 2020 dke.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machine

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"time"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/gmanager"
	"github.com/gostship/kunkka/pkg/provider/phases/clean"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	machineMaxReconciles = 2
)

// machineReconciler reconciles a machine object
type machineReconciler struct {
	client.Client
	Log    logr.Logger
	Mgr    manager.Manager
	Scheme *runtime.Scheme
	*gmanager.GManager
}

type manchineContext struct {
	Key    types.NamespacedName
	Logger logr.Logger
	*devopsv1.Cluster
	*devopsv1.Machine
	*devopsv1.ClusterCredential
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &machineReconciler{
		Client:   mgr.GetClient(),
		Mgr:      mgr,
		Log:      ctrl.Log.WithName("controllers").WithName("machine"),
		Scheme:   mgr.GetScheme(),
		GManager: pMgr,
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create machine controller")
	}

	return nil
}

func (r *machineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Machine{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: machineMaxReconciles}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.gostship.io,resources=machines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.gostship.io,resources=machines/status,verbs=get;update;patch

func (r *machineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("machine", req.NamespacedName.String())

	startTime := time.Now()
	defer func() {
		diffTime := time.Since(startTime)
		var logLevel klog.Level
		if diffTime > 1*time.Second {
			logLevel = 1
		} else if diffTime > 100*time.Millisecond {
			logLevel = 2
		} else {
			logLevel = 4
		}
		klog.V(logLevel).Infof("##### [%s] reconciling is finished. time taken: %v. ", req.NamespacedName.String(), diffTime)
	}()

	m := &devopsv1.Machine{}
	err := r.Client.Get(ctx, req.NamespacedName, m)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "not find machine")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get machine")
		return reconcile.Result{}, err
	}

	if !m.ObjectMeta.DeletionTimestamp.IsZero() {
		err := r.cleanMachinesResources(ctx, logger, m)
		if err != nil {
			logger.Error(err, "failed to clean machine resources")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if !constants.ContainsString(m.ObjectMeta.Finalizers, constants.FinalizersMachine) {
		logger.V(4).Info("start set", "finalizers", constants.FinalizersMachine)
		if m.ObjectMeta.Finalizers == nil {
			m.ObjectMeta.Finalizers = []string{}
		}
		m.ObjectMeta.Finalizers = append(m.ObjectMeta.Finalizers, constants.FinalizersMachine)
		err := r.Client.Update(ctx, m)
		if err != nil {
			logger.Error(err, "failed to set finalizers")
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	if m.Spec.Pause == true {
		logger.Info("machine is Pause")
		return reconcile.Result{}, nil
	}

	if len(string(m.Status.Phase)) == 0 {
		m.Status.Phase = devopsv1.MachineInitializing
		err = r.Client.Status().Update(ctx, m)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	cluster := &devopsv1.Cluster{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: m.Spec.ClusterName, Namespace: m.Namespace}, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "not find cluster")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get cluster")
		return reconcile.Result{}, err
	}

	if cluster.Status.Phase != devopsv1.ClusterRunning {
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: 30 * time.Second,
		}, nil
	}

	credential := &devopsv1.ClusterCredential{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: m.Spec.ClusterName, Namespace: m.Namespace}, credential)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("not find ClusterCredential")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get ClusterCredential")
		return reconcile.Result{}, err
	}

	klog.Infof("name: %s", cluster.Name)

	r.reconcile(ctx, &manchineContext{
		Key:               req.NamespacedName,
		Logger:            logger,
		Machine:           m,
		Cluster:           cluster,
		ClusterCredential: credential,
	})
	return ctrl.Result{}, nil
}

func (r *machineReconciler) cleanMachinesResources(ctx context.Context, logger logr.Logger, m *devopsv1.Machine) error {
	clusterCtx, err := r.ClusterManager.Get(m.Name)
	if err == nil {
		logger.Info("start delete node")
		clusterCtx.KubeCli.CoreV1().Nodes().Delete(ctx, m.Name, metav1.DeleteOptions{})
	}

	ssh, err := m.Spec.Machine.SSH()
	if err != nil {
		logger.Error(err, "failed new ssh")
		return err
	}

	logger.Info("start clean node")
	err = clean.CleanNode(ssh)
	if err != nil {
		logger.Error(err, "failed clean machine node")
		return err
	}

	logger.Info("start clean machine finalizers")
	m.ObjectMeta.Finalizers = constants.RemoveString(m.ObjectMeta.Finalizers, constants.FinalizersMachine)
	return r.Client.Update(ctx, m)
}
