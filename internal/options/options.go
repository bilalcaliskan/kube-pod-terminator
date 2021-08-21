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
	// MasterUrl is the cluster master ip to access
	MasterUrl string
	// InCluster is the if kube-pod-terminator is running in cluster or not
	InCluster bool
	// KubeConfigPath is the config file path to access cluster
	KubeConfigPath string
	// Namespace is the namespace of the kube-pod-terminator run on
	Namespace string
	// TickerIntervalMin is the Interval of scheduled job to run
	TickerIntervalMin int
	// ChannelCapacity is the capacity for concurrency
	ChannelCapacity int
	// GracePeriodSeconds is the grace period to delete pods
	GracePeriodSeconds int64
}

func (kpto *KubePodTerminatorOptions) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&kpto.MasterUrl, "masterUrl", "", "Cluster master ip to access")
	fs.BoolVar(&kpto.InCluster, "inCluster", true, "Specify if kube-pod-terminator is running in cluster")
	fs.StringVar(&kpto.KubeConfigPath, "kubeConfigPath", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"Kube config file path to access cluster. Required while running out of Kubernetes cluster")
	fs.StringVar(&kpto.Namespace, "namespace", "default", "Namespace to run on. Defaults to default namespace")
	fs.IntVar(&kpto.TickerIntervalMin, "tickerIntervalMin", 5, "Interval of scheduled job to run")
	fs.IntVar(&kpto.ChannelCapacity, "channelCapacity", 10, "Channel capacity for concurrency")
	fs.Int64Var(&kpto.GracePeriodSeconds, "gracePeriodSeconds", 30, "Grace period to delete pods")
}
