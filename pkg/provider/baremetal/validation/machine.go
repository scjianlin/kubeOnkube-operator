package validation

import (
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateMachine validates a given machine.
func ValidateMachine(machine *devopsv1.Machine) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, ValidateMachineSpec(&machine.Spec, field.NewPath("spec"))...)

	return allErrs
}

// ValidateMachineSpec validates a given machine spec.
func ValidateMachineSpec(spec *devopsv1.MachineSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// s, err := spec.SSH()
	// if err == nil {
	// 	if gpu.IsEnable(spec.Labels) {
	// 		if !gpu.MachineIsSupport(s) {
	// 			allErrs = append(allErrs, field.Invalid(fldPath.Child("labels"), spec.Labels, "must have GPU card if set GPU label"))
	// 		}
	// 	}
	// }

	return allErrs
}
