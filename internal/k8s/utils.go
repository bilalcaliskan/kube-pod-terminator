package k8s

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

// GetConfig gets parameters to generate rest.Config and returns it
func GetConfig(kubeConfigPath string, inCluster bool) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	if inCluster {
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
	} else {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// GetClientSet generates and returns k8s.Clientset using rest.Config
func GetClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func getTerminatingPods(ctx context.Context, clientSet kubernetes.Interface, namespace string) ([]v1.Pod, error) {
	var (
		resultSlice []v1.Pod
		pods        = new(v1.PodList)
		namespaces  = new(v1.NamespaceList)
		err         error
	)

	if strings.ToLower(namespace) == "all" {
		if namespaces, err = clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err != nil {
			return nil, err
		}
	} else {
		/*nsFieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", namespace))
		if namespaces, err = clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
			FieldSelector: nsFieldSelector.String(),
		}); err != nil {
			return nil, err
		}*/
		var ns *v1.Namespace
		if ns, err = clientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
			return nil, err
		}

		namespaces.Items = append(namespaces.Items, *ns)
	}

	for _, v := range namespaces.Items {
		var nsPods *v1.PodList
		if nsPods, err = clientSet.CoreV1().Pods(v.Name).List(ctx, metav1.ListOptions{}); err != nil {
			return nil, err
		}

		pods.Items = append(pods.Items, nsPods.Items...)
	}

	for _, pod := range pods.Items {
		deletionTimestamp := pod.ObjectMeta.DeletionTimestamp
		if deletionTimestamp != nil && deletionTimestamp.Add(time.Duration(opts.TerminatingStateMinutes)*time.Minute).Before(time.Now()) {
			resultSlice = append(resultSlice, pod)
		}
	}

	return resultSlice, nil
}

func getEvictedPods(ctx context.Context, clientSet kubernetes.Interface, namespace string) ([]v1.Pod, error) {
	var (
		evictedPods []v1.Pod
		pods        *v1.PodList
		namespaces  = new(v1.NamespaceList)
		err         error
	)

	if pods, err = clientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	}

	if strings.ToLower(namespace) == "all" {
		if namespaces, err = clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err != nil {
			return nil, err
		}
	} else {
		nsFieldSelector := fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", namespace))
		if namespaces, err = clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
			FieldSelector: nsFieldSelector.String(),
		}); err != nil {
			return nil, err
		}
	}

	for _, v := range namespaces.Items {
		var nsPods *v1.PodList
		if nsPods, err = clientSet.CoreV1().Pods(v.Name).List(ctx, metav1.ListOptions{}); err != nil {
			return nil, err
		}

		pods.Items = append(pods.Items, nsPods.Items...)
	}

	for _, pod := range pods.Items {
		if pod.Status.Reason == "Evicted" {
			evictedPods = append(evictedPods, pod)
		}
	}

	return evictedPods, nil
}
