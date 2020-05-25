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
	"flag"

	"github.com/gostship/kunkka/pkg/option"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string) *cobra.Command {
	opt := option.DefaultGlobalManagetOption()
	rootCmd := &cobra.Command{
		Use:               "mid-operator",
		Short:             "Request a new project",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Run:               runHelp,
	}

	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().StringVarP(&opt.Namespace, "namespace", "n", opt.Namespace, "Config namespace")
	rootCmd.PersistentFlags().BoolVarP(&opt.LoggerDevMode, "logger-dev-mode", "d", opt.LoggerDevMode, "Set development mode (mainly for logging)")
	rootCmd.PersistentFlags().IntVarP(&opt.GoroutineThreshold, "goroutine-threshold", "g", opt.GoroutineThreshold, "the max Goroutine Threshold")
	rootCmd.PersistentFlags().IntVarP(&opt.Threadiness, "threadiness", "t", opt.Threadiness, "the max Goroutine for controller reconcile")
	rootCmd.PersistentFlags().DurationVar(&opt.ResyncPeriod, "resync-period", opt.ResyncPeriod, "the max resync period to informer")

	// Make sure that klog logging variables are initialized so that we can
	// update them from this file.
	klog.InitFlags(nil)
	ctrl.SetLogger(zap.New(zap.UseDevMode(opt.LoggerDevMode)))

	// Make sure klog (used by the client-go dependency) logs to stderr, as it
	// will try to log to directories that may not exist in the cilium-operator
	// container (/tmp) and cause the cilium-operator to exit.
	flag.Set("logtostderr", "true")
	AddFlags(rootCmd)

	rootCmd.AddCommand(NewControllerCmd(opt))
	rootCmd.AddCommand(NewCmdVersion())
	return rootCmd
}

func hideInheritedFlags(orig *cobra.Command, hidden ...string) {
	orig.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		for _, hidden := range hidden {
			_ = cmd.Flags().MarkHidden(hidden) // nolint: errcheck
		}

		orig.SetHelpFunc(nil)
		orig.HelpFunc()(cmd, args)
	})
}
