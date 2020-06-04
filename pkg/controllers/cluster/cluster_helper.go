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
	"time"

	"github.com/gostship/kunkka/pkg/provider"
	clusterprovider "github.com/gostship/kunkka/pkg/provider/cluster"
)

const (
	clusterClientRetryCount    = 5
	clusterClientRetryInterval = 5 * time.Second

	reasonFailedInit   = "FailedInit"
	reasonFailedUpdate = "FailedUpdate"
)

func (r *clusterReconciler) onCreate(ctx *reconcileContext) error {
	p, err := clusterprovider.GetProvider(ctx.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper, err := provider.GetCluster(ctx.Ctx, r.Client, ctx.Cluster)
	if err != nil {
		return err
	}
	err = p.OnCreate(ctx.Ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Status.Message = err.Error()
		clusterWrapper.Status.Reason = reasonFailedInit
		r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
		return err
	}

	return nil
}

func (r *clusterReconciler) onUpdate(ctx *reconcileContext) error {
	p, err := clusterprovider.GetProvider(ctx.Cluster.Spec.Type)
	if err != nil {
		return err
	}

	clusterWrapper, err := provider.GetCluster(ctx.Ctx, r.Client, ctx.Cluster)
	if err != nil {
		return err
	}

	err = p.OnUpdate(ctx.Ctx, clusterWrapper)
	if err != nil {
		clusterWrapper.Status.Message = err.Error()
		clusterWrapper.Status.Reason = reasonFailedUpdate
		r.Client.Status().Update(ctx.Ctx, ctx.Cluster)
		return err
	}
	clusterWrapper.Status.Message = ""
	clusterWrapper.Status.Reason = ""
	r.Client.Status().Update(ctx.Ctx, clusterWrapper.ClusterCredential)
	r.Client.Status().Update(ctx.Ctx, clusterWrapper.Cluster)
	return nil
}
