package app

import (
	apiManager "github.com/gostship/kunkka/pkg/apimanager"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/k8sclient"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"time"
)

var (
	logger = logf.KBLog.WithName("admin-api")
)

// returns NewAPICmd
func NewAPICmd(cli *KunkkaCli) *cobra.Command {
	opt := apiManager.DefaultOption()
	cmd := &cobra.Command{
		Use:     "api",
		Aliases: []string{"api"},
		Short:   "Manage kunkka api server",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())

			cfg, err := cli.GetK8sConfig()
			if err != nil {
				klog.Fatalf("unable to get kubeconfig err: %v", err)
			}

			rp := time.Second * 120
			mgr, err := ctrlmanager.New(cfg, ctrlmanager.Options{
				Scheme:             k8sclient.GetScheme(),
				MetricsBindAddress: "0",
				LeaderElection:     false,
				SyncPeriod:         &rp,
			})
			if err != nil {
				klog.Fatalf("unable to new kunkka manager err: %v", err)
			}

			stopCh := signals.SetupSignalHandler()

			k8sCli := k8smanager.MasterClient{
				KubeCli: cli.GetKubeInterfaceOrDie(),
				Manager: mgr,
			}

			apiMgr, err := apiManager.NewAPIManager(k8sCli, opt, "controller")
			if err != nil {
				klog.Fatalf("unable to NewKunkkaApiManager err: %v", err)
			}

			// add http server Runnable
			mgr.Add(apiMgr.Router)

			// add k8s cluster manager Runnable
			//mgr.Add(apiMgr.K8sMgr)

			logger.Info("zap debug", "SyncPeriod", rp)
			klog.Info("starting manager")
			if err := mgr.Start(stopCh); err != nil {
				klog.Fatalf("problem start running manager err: %v", err)
			}
		},
	}

	cmd.PersistentFlags().IntVar(&opt.GoroutineThreshold, "goroutine-threshold", opt.GoroutineThreshold, "the max Goroutine Threshold")
	cmd.PersistentFlags().StringVar(&opt.HTTPAddr, "http-addr", opt.HTTPAddr, "HttpAddr for some info")
	cmd.PersistentFlags().BoolVar(&opt.IsMeta, "is-meta", opt.IsMeta, "Whether it is a meta cluster")
	cmd.PersistentFlags().BoolVar(&opt.GinLogEnabled, "enable-ginlog", opt.GinLogEnabled, "Enabled will open gin run log.")
	cmd.PersistentFlags().BoolVar(&opt.PprofEnabled, "enable-pprof", opt.PprofEnabled, "Enabled will open endpoint for go pprof.")
	return cmd
}

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
