package app_option

import (
	"github.com/gostship/kunkka/pkg/option"
	"github.com/spf13/pflag"
)

type Options struct {
	Global *option.GlobalManagerOption
	Ctrl   *option.ControllersManagerOption
}

// NewOptions creates a new Options with a default config.
func NewOptions() *Options {
	return &Options{
		Global: option.DefaultGlobalManagetOption(),
		Ctrl:   option.DefaultControllersManagerOption(),
	}
}

// AddFlags adds flags for a specific server to the specified FlagSet object.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.Global.AddFlags(fs)
	o.Ctrl.AddFlags(fs)
}
