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

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// manchineReconciler reconciles a machine object
type manchineReconciler struct {
	client.Client
	Log    logr.Logger
	Mgr    manager.Manager
	Scheme *runtime.Scheme
}

func Add(mgr manager.Manager) error {
	reconciler := &manchineReconciler{
		Client: mgr.GetClient(),
		Mgr:    mgr,
		Log:    ctrl.Log.WithName("controllers").WithName("machine"),
		Scheme: mgr.GetScheme(),
	}

	err := reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrapf(err, "unable to create machine controller")
	}

	return nil
}

func (r *manchineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devopsv1.Machine{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=devops.gostship.io,resources=virtulclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=devops.gostship.io,resources=virtulclusters/status,verbs=get;update;patch

func (r *manchineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	_ = context.Background()
	_ = r.Log.WithValues("cluster", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}