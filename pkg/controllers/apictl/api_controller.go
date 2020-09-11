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

package apictl

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/controllers/common"
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

// apiReconciler reconciles a Cluster object
type apiReconciler struct {
	client.Client
	*gmanager.GManager
	Log            logr.Logger
	Mgr            manager.Manager
	Scheme         *runtime.Scheme
	ClusterStarted map[string]bool
}

type apiContext struct {
	Key     types.NamespacedName
	Logger  logr.Logger
	Cluster *devopsv1.Cluster
}

func Add(mgr manager.Manager, pMgr *gmanager.GManager) error {
	reconciler := &apiReconciler{
		Client:         mgr.GetClient(),
		Mgr:            mgr,
		Log:            ctrl.Log.WithName("api-controllers").WithName("api-controllers"),
		Scheme:         mgr.GetScheme(),
		GManager:       pMgr,
		ClusterStarted: make(map[string]bool),
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create api controller")
	}

	return nil
}

func (r *apiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Cluster{}).
		Owns(&devopsv1.ClusterCredential{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.gostship.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.gostship.io,resources=clusters/status,verbs=get;update;patch
func (r *apiReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("api-controller", req.NamespacedName.String())

	startTime := time.Now()
	defer func() {
		diffTime := time.Since(startTime)
		var logLevel klog.Level
		if diffTime > 1*time.Second {
			logLevel = 2
		} else if diffTime > 100*time.Millisecond {
			logLevel = 4
		} else {
			logLevel = 5
		}
		klog.V(logLevel).Infof("##### [%s] reconciling manager cluster client is finished. time taken: %v. ", req.NamespacedName.String(), diffTime)
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

	rc := &apiContext{
		Key:     req.NamespacedName,
		Logger:  logger,
		Cluster: c,
	}

	if !c.ObjectMeta.DeletionTimestamp.IsZero() {
		err := r.ClusterManager.Delete(c.Name)
		if err != nil {
			logger.Error(err, "failed to clean cluster client resources")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	r.reconcile(ctx, rc)
	return ctrl.Result{}, nil
}

func (r *apiReconciler) addClusterCheck(ctx context.Context, c *common.Cluster) error {
	if _, ok := r.ClusterStarted[c.Cluster.Name]; ok {
		if extKubeconfig, ok := c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName]; ok {
			klog.V(4).Infof("cluster: %s, add manager success!", c.Cluster.Name)
			cls, err := r.GManager.AddNewClusters(c.Cluster.Name, extKubeconfig)
			if err != nil {
				klog.Errorf("failed add cluster client: %s manager cache", c.Cluster.Name)
				return nil
			}

			klog.Infof("#######  add cluster client: %s to manager cache success", c.Cluster.Name)
			err = r.GManager.Update(cls)
			if err != nil {
				klog.Errorf("failed update cluster client: %s manager cache", c.Cluster.Name)
				return nil
			}

			return nil
		}
	} else {
		if extKubeconfig, ok := c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName]; ok {
			klog.V(4).Infof("cluster client: %s, add manager success!", c.Cluster.Name)
			_, err := r.GManager.AddNewClusters(c.Cluster.Name, extKubeconfig)
			if err != nil {
				klog.Errorf("failed add cluster client: %s manager cache", c.Cluster.Name)
				return nil
			}

			klog.Infof("#######  add cluster client: %s to manager cache success", c.Cluster.Name)
			r.ClusterStarted[c.Cluster.Name] = true
			return nil
		}
	}

	klog.Warningf("can't find %s", pkiutil.ExternalAdminKubeConfigFileName)
	return nil
}

func (r *apiReconciler) reconcile(ctx context.Context, rc *apiContext) error {

	clusterWrapper, err := common.GetCluster(ctx, r.Client, rc.Cluster, r.ClusterManager)
	if err != nil {
		return err
	}

	switch rc.Cluster.Status.Phase {
	case devopsv1.ClusterInitializing:
		rc.Logger.Info("onCreate")
	case devopsv1.ClusterRunning:
		rc.Logger.Info("onUpdate")
		r.addClusterCheck(ctx, clusterWrapper)
	default:
		return fmt.Errorf("no handler for %q", rc.Cluster.Status.Phase)
	}
	return nil
}
