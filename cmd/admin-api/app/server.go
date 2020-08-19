package app

import (
	"flag"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// returns the api server of cobra
func GetApiCmd(args []string) *cobra.Command {
	opt := DefaultRootOption()

	//apiOpt := NewOptions()

	apicmd := &cobra.Command{
		Use:               "kunkka-api",
		Short:             "Request a new api server",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		Run:               runHelp,
	}

	apicmd.SetArgs(args)

	// Make sure that klog logging variables are initialized so that we can
	klog.InitFlags(nil)
	logf.SetLogger(logf.ZapLogger(opt.DevelopmentMode))

	// Make sure klog (used by the client-go dependency) logs to stderr, as it
	// will try to log to directories that may not exist in the cilium-operator
	// container (/tmp) and cause the cilium-operator to exit.
	flag.Set("logtostderr", "true")

	AddFlags(apicmd)
	cli := NewKunkkaCli(opt)
	apicmd.AddCommand(NewAPICmd(cli))
	apicmd.AddCommand(NewCmdVersion(cli))

	return apicmd
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// AddFlags function is add flag
func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}
