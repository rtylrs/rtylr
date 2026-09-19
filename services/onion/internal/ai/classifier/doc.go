// Package classifier scores form elements against candidate document values using
// the CUA-S1-FORMS inference path (trycua/cua, libs/cua-s1, MIT), ported to
// Go in the checkpoint and model subpackages. classifier_test.go and real_weights_test.go prove parity
// against the PyTorch reference; scope is encoder tinyx, F32 weights, CPU.
package classifier
