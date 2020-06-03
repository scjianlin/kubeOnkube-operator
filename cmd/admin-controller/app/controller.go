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
package app

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/gostship/kunkka/cmd/admin-controller/app/app_option"
	"github.com/gostship/kunkka/pkg/controllers"
	"github.com/gostship/kunkka/pkg/k8sclient"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func NewControllerCmd(opt *app_option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ctrl",
		Short: "Manage controller Component",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())

			cfg, err := opt.Global.GetK8sConfig()
			if err != nil {
				klog.Fatalf("unable to get cfg err: %v", err)
			}

			// Adjust our client's rate limits based on the number of controllers we are running.
			cfg.QPS = float32(2) * cfg.QPS
			cfg.Burst = 2 * cfg.Burst

			mgr, err := ctrlmanager.New(cfg, ctrlmanager.Options{
				Scheme:                  k8sclient.GetScheme(),
				LeaderElection:          opt.Global.EnableLeaderElection,
				LeaderElectionNamespace: opt.Global.LeaderElectionNamespace,
				SyncPeriod:              &opt.Global.ResyncPeriod,
				MetricsBindAddress:      "0",
				HealthProbeBindAddress:  ":8090",
				// Port:               9443,
			})
			if err != nil {
				klog.Fatalf("unable to new manager err: %v", err)
			}

			// crds, err := static.LoadCRDs()
			// if err != nil {
			// 	klog.Fatalf("unable to get cfg err: %v", err)
			// }
			//
			// for _, crd := range crds {
			// 	crderr := k8sutils.Reconcile(log.Log, mgr.GetClient(), crd, k8sutils.DesiredStatePresent)
			// 	if crderr != nil {
			// 		klog.Errorf("crd name: %s err: %v", crd.Name, crderr)
			// 	}
			// }

			// Setup all Controllers
			klog.Info("Setting up controller")
			if err := controllers.AddToManager(mgr, opt.Ctrl); err != nil {
				klog.Fatalf("unable to register controllers to the manager err: %v", err)
			}

			klog.Info("starting manager")
			stopCh := signals.SetupSignalHandler()
			if err := mgr.Start(stopCh); err != nil {
				klog.Fatalf("problem start running manager err: %v", err)
			}
		},
	}

	opt.Ctrl.AddFlags(cmd.Flags())
	return cmd
}
