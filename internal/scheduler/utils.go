package scheduler

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// GetConfig gets parameters to generate rest.Config and returns it
func GetConfig(kubeConfigPath string, inCluster bool) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}

// GetClientSet generates and returns kubernetes.Clientset using rest.Config
func GetClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func getTerminatingPods(ctx context.Context, clientSet *kubernetes.Clientset, namespace string) ([]v1.Pod, error) {
	var resultSlice []v1.Pod
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		deletionTimestamp := pod.ObjectMeta.DeletionTimestamp
		if deletionTimestamp != nil && deletionTimestamp.Add(time.Duration(opts.TerminatingPodThreshold)*time.Minute).Before(time.Now()) {
			resultSlice = append(resultSlice, pod)
		}
	}
	return resultSlice, nil
}

func getEvictedPods(ctx context.Context, clientSet *kubernetes.Clientset, namespace string) ([]v1.Pod, error) {
	var evictedPods []v1.Pod
	pods, err := clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		// if pod.Status.Reason == "Evicted" {
		// if pod.Name == "cronjob-sample-1635125520-r6v5c" {
		// logger.Info("fetched pod", zap.Any("pod", pod))
		// }

		if pod.Status.Phase == "Pending" {
			evictedPods = append(evictedPods, pod)
		}
	}

	return evictedPods, nil
}
