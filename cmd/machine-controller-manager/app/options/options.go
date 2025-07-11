/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file was copied and modified from the kubernetes/kubernetes project
https://github.com/kubernetes/kubernetes/release-1.8/cmd/kube-controller-manager/app/options/options.go

Modifications Copyright SAP SE or an SAP affiliate company and Gardener contributors
*/

package options

import (
	"fmt"
	"mime"
	"net"
	"time"

	machineconfig "github.com/gardener/machine-controller-manager/pkg/options"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/component-base/logs"

	"github.com/gardener/machine-controller-manager/pkg/util/client/leaderelectionconfig"

	// add the machine feature gates
	"github.com/gardener/machine-controller-manager/pkg/apis/constants"
	_ "github.com/gardener/machine-controller-manager/pkg/features"
)

// MCMServer is the main context object for the controller manager.
type MCMServer struct {
	machineconfig.MachineControllerManagerConfiguration

	ControlKubeconfig string
	TargetKubeconfig  string
}

// NewMCMServer creates a new MCMServer with a default config.
func NewMCMServer() *MCMServer {

	s := MCMServer{
		// Part of these default values also present in 'cmd/cloud-controller-manager/app/options/options.go'.
		// Please keep them in sync when doing update.
		MachineControllerManagerConfiguration: machineconfig.MachineControllerManagerConfiguration{
			Port:                    10258,
			Namespace:               "default",
			Address:                 "0.0.0.0",
			ConcurrentNodeSyncs:     10,
			ContentType:             "application/vnd.kubernetes.protobuf",
			MinResyncPeriod:         metav1.Duration{Duration: 12 * time.Hour},
			KubeAPIQPS:              20.0,
			KubeAPIBurst:            30,
			LeaderElection:          leaderelectionconfig.DefaultLeaderElectionConfiguration(),
			ControllerStartInterval: metav1.Duration{Duration: 0 * time.Second},
			AutoscalerScaleDownAnnotationDuringRollout: true,
			SafetyOptions: machineconfig.SafetyOptions{
				SafetyUp:                        2,
				SafetyDown:                      1,
				MachineSafetyOvershootingPeriod: metav1.Duration{Duration: 1 * time.Minute},
			},
		},
	}
	s.LeaderElection.LeaderElect = true
	return &s
}

// AddFlags adds flags for a specific CMServer to the specified FlagSet
func (s *MCMServer) AddFlags(fs *pflag.FlagSet) {
	fs.Int32Var(&s.Port, "port", s.Port, "The port that the controller-manager's http service runs on")
	fs.Var(machineconfig.IPVar{Val: &s.Address}, "address", "The IP address to serve on (set to 0.0.0.0 for all interfaces)")
	fs.StringVar(&s.CloudProvider, "cloud-provider", s.CloudProvider, "The provider for cloud services.  Empty string for no provider.")
	fs.Int32Var(&s.ConcurrentNodeSyncs, "concurrent-syncs", s.ConcurrentNodeSyncs, "The number of nodes that are allowed to sync concurrently. Larger number = more responsive service management, but more CPU (and network) load")
	fs.DurationVar(&s.MinResyncPeriod.Duration, "min-resync-period", s.MinResyncPeriod.Duration, "The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod")
	fs.BoolVar(&s.EnableProfiling, "profiling", false, "Enable profiling via web interface host:port/debug/pprof/")
	fs.BoolVar(&s.EnableContentionProfiling, "contention-profiling", false, "Enable lock contention profiling, if profiling is enabled")
	fs.StringVar(&s.TargetKubeconfig, "target-kubeconfig", s.TargetKubeconfig, fmt.Sprintf("Filepath to the target cluster's kubeconfig where node objects are expected to join or %q if there is no target cluster", constants.TargetKubeconfigDisabledValue))
	fs.StringVar(&s.ControlKubeconfig, "control-kubeconfig", s.ControlKubeconfig, "Filepath to the control cluster's kubeconfig where machine objects would be created. Optionally you could also use 'inClusterConfig' when pod is running inside control kubeconfig. (Default value is same as target-kubeconfig)")
	fs.StringVar(&s.Namespace, "namespace", s.Namespace, "Name of the namespace in control cluster where controller would look for CRDs and Kubernetes objects")
	fs.StringVar(&s.ContentType, "kube-api-content-type", s.ContentType, "Content type of requests sent to apiserver.")
	fs.Float32Var(&s.KubeAPIQPS, "kube-api-qps", s.KubeAPIQPS, "QPS to use while talking with kubernetes apiserver")
	fs.Int32Var(&s.KubeAPIBurst, "kube-api-burst", s.KubeAPIBurst, "Burst to use while talking with kubernetes apiserver")
	fs.DurationVar(&s.ControllerStartInterval.Duration, "controller-start-interval", s.ControllerStartInterval.Duration, "Interval between starting controller managers.")

	fs.Int32Var(&s.SafetyOptions.SafetyUp, "safety-up", s.SafetyOptions.SafetyUp, "The number of excess machine objects permitted for any machineSet/machineDeployment beyond its expected number of replicas based on desired and max-surge, we call this the upper-limit. When this upper-limit is reached, the objects are temporarily frozen until the number of objects reduce. upper-limit = desired + maxSurge (if applicable) + safetyUp.")
	fs.Int32Var(&s.SafetyOptions.SafetyDown, "safety-down", s.SafetyOptions.SafetyDown, "Upper-limit minus safety-down value gives the lower-limit. This is the limits below which any temporarily frozen machineSet/machineDeployment object is unfrozen. lower-limit = desired + maxSurge (if applicable) + safetyUp - safetyDown.")

	fs.DurationVar(&s.SafetyOptions.MachineSafetyOvershootingPeriod.Duration, "machine-safety-overshooting-period", s.SafetyOptions.MachineSafetyOvershootingPeriod.Duration, "Time period (in duration) used to poll for overshooting of machine objects backing a machineSet by safety controller.")

	fs.BoolVar(&s.AutoscalerScaleDownAnnotationDuringRollout, "autoscaler-scaledown-annotation-during-rollout", true, "Add cluster autoscaler scale-down disabled annotation during roll-out.")

	logs.AddFlags(fs) // Here `logs` is `k8s.io/component-base/logs`.

	leaderelectionconfig.BindFlags(&s.LeaderElection, fs)
	// TODO: DefaultFeatureGate is global and it adds all k8s flags
	// utilfeature.DefaultFeatureGate.AddFlag(fs)
}

// Validate is used to validate the options and config before launching the controller manager
func (s *MCMServer) Validate() error {
	var errs []error
	if s.Port < 1 || s.Port > 65535 {
		errs = append(errs, fmt.Errorf("invalid port number provided: got %d", s.Port))
	}
	if ip := net.ParseIP(s.Address); ip == nil {
		errs = append(errs, fmt.Errorf("invalid IP address provided: got: %v", ip))
	}
	if s.ConcurrentNodeSyncs <= 0 {
		errs = append(errs, fmt.Errorf("concurrent syncs should be greater than zero: got: %d", s.ConcurrentNodeSyncs))
	}
	if s.MinResyncPeriod.Duration < 0 {
		errs = append(errs, fmt.Errorf("min resync period should be a non negative value: got: %v", s.MinResyncPeriod.Duration))
	}
	if !s.EnableProfiling && s.EnableContentionProfiling {
		errs = append(errs, fmt.Errorf("contention-profiling cannot be enabled without enabling profiling"))
	}
	if _, _, err := mime.ParseMediaType(s.ContentType); err != nil {
		errs = append(errs, fmt.Errorf("kube api content type cannot be parsed: %w", err))
	}
	if s.KubeAPIQPS <= 0 {
		errs = append(errs, fmt.Errorf("kube api qps should be greater than zero: got: %f", s.KubeAPIQPS))
	}
	if s.KubeAPIBurst < 0 {
		errs = append(errs, fmt.Errorf("kube api burst should not be a negative value: got: %d", s.KubeAPIBurst))
	}
	if s.ControllerStartInterval.Duration < 0 {
		errs = append(errs, fmt.Errorf("controller start interval should be a non negative value: got: %v", s.ControllerStartInterval.Duration))
	}
	if s.SafetyOptions.SafetyUp < 0 {
		errs = append(errs, fmt.Errorf("safety up should be a non negative value: got: %d", s.SafetyOptions.SafetyUp))
	}
	if s.SafetyOptions.SafetyDown < 0 {
		errs = append(errs, fmt.Errorf("safety down should be a non negative value: got: %d", s.SafetyOptions.SafetyDown))
	}
	if s.SafetyOptions.MachineSafetyOvershootingPeriod.Duration < 0 {
		errs = append(errs, fmt.Errorf("machine safety overshooting period should be a non negative number: got: %v", s.SafetyOptions.MachineSafetyOvershootingPeriod.Duration))
	}
	if s.ControlKubeconfig == "" && s.TargetKubeconfig == constants.TargetKubeconfigDisabledValue {
		errs = append(errs, fmt.Errorf("--control-kubeconfig cannot be empty if --target-kubeconfig=%s is specified", constants.TargetKubeconfigDisabledValue))
	}

	return utilerrors.NewAggregate(errs)
}
