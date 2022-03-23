package options

import (
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
)

var kubePodTerminatorOptions = &KubePodTerminatorOptions{}

func init() {
	kubePodTerminatorOptions.addFlags(pflag.CommandLine)
	pflag.Parse()
}

// GetKubePodTerminatorOptions returns the pointer of SynFloodOptions
func GetKubePodTerminatorOptions() *KubePodTerminatorOptions {
	return kubePodTerminatorOptions
}

// KubePodTerminatorOptions contains frequent command line and application options.
type KubePodTerminatorOptions struct {
	// InCluster is the if kube-pod-terminator is running in cluster or not
	InCluster bool
	// KubeConfigPaths is the comma separated list of kubeconfig file paths to access with the cluster
	KubeConfigPaths string
	// Namespace is the namespace of the kube-pod-terminator run on
	Namespace string
	// TickerIntervalMinutes is the Interval of scheduled job to run
	TickerIntervalMinutes int32
	// ChannelCapacity is the capacity for concurrency
	ChannelCapacity int
	// GracePeriodSeconds is the grace period to delete pods
	GracePeriodSeconds int64
	// TerminateEvicted is a boolean flag to tell if terminating evicted pods is supported
	TerminateEvicted bool
	// TerminatingStateMinutes is the specifier to select pods which are more in terminating state
	TerminatingStateMinutes int32
	// OneShot is the specifier to run kube-pod-terminator only one time instead of continuously running in the background
	OneShot bool
}

func (kpto *KubePodTerminatorOptions) addFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&kpto.InCluster, "inCluster", true, "Specify if kube-pod-terminator is running in cluster")
	fs.StringVar(&kpto.KubeConfigPaths, "kubeConfigPaths", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"comma separated list of kubeconfig file paths to access with the cluster")
	fs.StringVar(&kpto.Namespace, "namespace", "all", "Namespace to run on. Defaults to default namespace")
	fs.Int32Var(&kpto.TickerIntervalMinutes, "tickerIntervalMinutes", 5, "Interval of scheduled job to run")
	fs.IntVar(&kpto.ChannelCapacity, "channelCapacity", 10, "Channel capacity for concurrency")
	fs.Int64Var(&kpto.GracePeriodSeconds, "gracePeriodSeconds", 30, "Grace period to delete pods")
	fs.BoolVar(&kpto.TerminateEvicted, "terminateEvicted", true, "Terminate evicted pods in specified namespaces")
	fs.Int32Var(&kpto.TerminatingStateMinutes, "terminatingStateMinutes", 30, "Terminate stucked pods "+
		"in terminating state which are more than that value")
	fs.BoolVar(&kpto.OneShot, "oneShot", false, "OneShot is the specifier to run kube-pod-terminator "+
		"only one time instead of continuously running in the background")
}
