package options

var kubePodTerminatorOptions = &KubePodTerminatorOptions{}

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
	// BannerFilePath is the relative path to the banner file
	BannerFilePath string
	// VerboseLog is the verbosity of the logging library
	VerboseLog bool
}
