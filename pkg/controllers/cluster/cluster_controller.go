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

package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/gmanager"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// clusterReconciler reconciles a Cluster object
type clusterReconciler struct {
	client.Client
	*gmanager.GManager
	Log            logr.Logger
	Mgr            manager.Manager
	Scheme         *runtime.Scheme
	ClusterStarted map[string]bool
}

type clusterContext struct {
	Key     types.NamespacedName
	Logger  logr.Logger
	Cluster *devopsv1.Cluster
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &clusterReconciler{
		Client:         mgr.GetClient(),
		Mgr:            mgr,
		Log:            ctrl.Log.WithName("controllers").WithName("cluster"),
		Scheme:         mgr.GetScheme(),
		GManager:       pMgr,
		ClusterStarted: make(map[string]bool),
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create cluster controller")
	}

	return nil
}

func (r *clusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Cluster{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.gostship.io,resources=virtulclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.gostship.io,resources=virtulclusters/status,verbs=get;update;patch

func (r *clusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("cluster", req.NamespacedName.Name)

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
		klog.V(logLevel).Infof("##### [%s] reconciling is finished. time taken: %v. ", req.NamespacedName, diffTime)
	}()

	c := &devopsv1.Cluster{}
	err := r.Client.Get(ctx, req.NamespacedName, c)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(4).Info("not find cluster")
			return reconcile.Result{}, nil
		}

		logger.Error(err, "failed to get cluster")
		return reconcile.Result{}, err
	}

	if c.Spec.Pause == true {
		logger.V(4).Info("cluster is Pause")
		return reconcile.Result{}, nil
	}

	klog.Infof("name: %s", c.Name)
	if len(string(c.Status.Phase)) == 0 {
		c.Status.Phase = devopsv1.ClusterInitializing
		err = r.Client.Status().Update(ctx, c)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	r.reconcile(ctx, &clusterContext{
		Key:     req.NamespacedName,
		Logger:  logger,
		Cluster: c,
	})
	return ctrl.Result{}, nil
}

func (r *clusterReconciler) addClusterCheck(ctx context.Context, rc *clusterContext) error {
	if _, ok := r.ClusterStarted[rc.Cluster.Name]; ok {
		return nil
	}
	cfgMap := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: rc.Cluster.Namespace, Name: "kube-apiserver-config"}, cfgMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Warningf("not find kube-apiserver-config cluster: %s", rc.Cluster.Name)
			return nil
		}
		klog.Warningf("failed get kube-apiserver-config cluster: %s", rc.Cluster.Name)
		return errors.Wrapf(err, "failed get kubeconfig cfgMap")
	}

	if extKubeconfig, ok := cfgMap.Data[pkiutil.ExternalAdminKubeConfigFileName]; ok {
		_, err := r.GManager.AddNewClusters(rc.Cluster.Name, extKubeconfig)
		if err != nil {
			klog.Errorf("failed add cluster: %s manager cache", rc.Cluster.Name)
			return nil
		}
		klog.Infof("#######  add cluster: %s to manager cache success", rc.Cluster.Name)
		r.ClusterStarted[rc.Cluster.Name] = true
		return nil
	}

	klog.Warningf("can't find %s", pkiutil.ExternalAdminKubeConfigFileName)
	return nil
}

func (r *clusterReconciler) reconcile(ctx context.Context, rc *clusterContext) error {
	switch rc.Cluster.Status.Phase {
	case devopsv1.ClusterInitializing:
		rc.Logger.Info("onCreate")
		return r.onCreate(ctx, rc)
	case devopsv1.ClusterRunning:
		rc.Logger.Info("onUpdate")
		r.addClusterCheck(ctx, rc)
		return r.onUpdate(ctx, rc)
	default:
		return fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}
}
