package k8s

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/bilalcaliskan/kube-pod-terminator/internal/logging"
	"github.com/bilalcaliskan/kube-pod-terminator/internal/options"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type FakeAPI struct {
	ClientSet kubernetes.Interface
	Namespace string
}

func getDefaultOpts() *options.KubePodTerminatorOptions {
	return &options.KubePodTerminatorOptions{
		InCluster:               true,
		KubeConfigPaths:         filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		Namespace:               "all",
		TickerIntervalMinutes:   5,
		GracePeriodSeconds:      30,
		TerminateEvicted:        true,
		TerminatingStateMinutes: 30,
		OneShot:                 false,
		BannerFilePath:          "",
		VerboseLog:              false,
	}
}

func (fAPI *FakeAPI) createEvictedPod(name, namespace string) (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			DeletionTimestamp: nil,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "varnish",
					Image:           "varnish:7.0.1",
					ImagePullPolicy: "Always",
					Ports: []v1.ContainerPort{
						{Name: "port1", ContainerPort: 6082, Protocol: v1.ProtocolTCP},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Reason: "Evicted",
		},
	}

	pod, err := fAPI.ClientSet.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (fAPI *FakeAPI) createNamespace(name string) (*v1.Namespace, error) {
	namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	namespace, err := fAPI.ClientSet.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

func (fAPI *FakeAPI) createTerminatingPod(name, namespace string, deletionTimestamp *metav1.Time) (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			DeletionTimestamp: deletionTimestamp,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "varnish",
					Image:           "varnish:7.0.1",
					ImagePullPolicy: "Always",
					Ports: []v1.ContainerPort{
						{Name: "port1", ContainerPort: 6082, Protocol: v1.ProtocolTCP},
					},
				},
			},
		},
	}

	pod, err := fAPI.ClientSet.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func getFakeAPI() *FakeAPI {
	client := fake.NewSimpleClientset()
	api := &FakeAPI{ClientSet: client, Namespace: "default"}
	return api
}

func TestRunNoTargetTerminatingPods(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	testOpts := getDefaultOpts()

	namespace := "default"
	_, _ = api.createNamespace(namespace)

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:          "case1",
			podName:           "varnish-pod-1",
			namespace:         namespace,
			deletionTimestamp: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createTerminatingPod(tc.podName, tc.namespace, tc.deletionTimestamp)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestRunDoNotTerminateEvictedPods(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	_, _ = api.createNamespace("default")

	testOpts := getDefaultOpts()
	testOpts.TerminateEvicted = false
	testOpts.Namespace = "default"
	pod, err := api.createTerminatingPod("varnish-pod-1", "default",
		&metav1.Time{Time: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC)})
	pod2, err2 := api.createEvictedPod("varnish-pod-2", "default")
	assert.Nil(t, err)
	assert.NotNil(t, pod)
	assert.Nil(t, err2)
	assert.NotNil(t, pod2)
	time.Sleep(2 * time.Second)

	Run(testOpts, api.ClientSet, "")
}

func TestRunEvictedPodsAllNamespaces(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	testOpts := getDefaultOpts()
	testOpts.Namespace = "all"
	_, _ = api.createNamespace("default")
	_, _ = api.createNamespace("kube-system")

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:  "case1",
			podName:   "varnish-pod-1",
			namespace: "default",
		},
		{
			caseName:  "case2",
			podName:   "varnish-pod-2",
			namespace: "kube-system",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createEvictedPod(tc.podName, tc.namespace)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestRunEvictedPodsAllNamespacesOneShot(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	testOpts := getDefaultOpts()
	testOpts.OneShot = true
	testOpts.Namespace = "all"
	_, _ = api.createNamespace("default")
	_, _ = api.createNamespace("kube-system")

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:  "case1",
			podName:   "varnish-pod-1",
			namespace: "default",
		},
		{
			caseName:  "case2",
			podName:   "varnish-pod-2",
			namespace: "kube-system",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createEvictedPod(tc.podName, tc.namespace)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestRunEvictedPodsSingleNamespace(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	_, _ = api.createNamespace("default")
	testOpts := getDefaultOpts()
	testOpts.Namespace = "default"

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:  "case1",
			podName:   "varnish-pod-1",
			namespace: "default",
		},
		{
			caseName:  "case2",
			podName:   "varnish-pod-2",
			namespace: "default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createEvictedPod(tc.podName, tc.namespace)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestRunBrokenApiCall(t *testing.T) {
	restConfig, err := GetConfig("../../mock/kubeconfig", false)
	assert.Nil(t, err)
	assert.NotNil(t, restConfig)

	testOpts := getDefaultOpts()
	testOpts.Namespace = "default"

	clientSet, err := GetClientSet(restConfig)
	assert.Nil(t, err)
	assert.NotNil(t, clientSet)

	Run(testOpts, clientSet, "")
}

func TestRunTerminatingPodsAllNamespaces(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	testOpts := getDefaultOpts()
	testOpts.Namespace = "all"

	_, _ = api.createNamespace("default")
	_, _ = api.createNamespace("kube-system")

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:          "case1",
			podName:           "varnish-pod-1",
			namespace:         "default",
			deletionTimestamp: &metav1.Time{Time: time.Date(2019, time.Month(2), 21, 1, 10, 30, 0, time.UTC)},
		},
		{
			caseName:          "case2",
			podName:           "varnish-pod-2",
			namespace:         "kube-system",
			deletionTimestamp: &metav1.Time{Time: time.Date(2019, time.Month(2), 21, 1, 10, 30, 0, time.UTC)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createTerminatingPod(tc.podName, tc.namespace, tc.deletionTimestamp)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestRunTerminatingPodsSingleNamespace(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	testOpts := getDefaultOpts()
	testOpts.Namespace = "default"

	_, _ = api.createNamespace("default")
	_, _ = api.createNamespace("kube-system")

	cases := []struct {
		caseName, podName, namespace string
		deletionTimestamp            *metav1.Time
	}{
		{
			caseName:          "case1",
			podName:           "varnish-pod-1",
			namespace:         "default",
			deletionTimestamp: &metav1.Time{Time: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createTerminatingPod(tc.podName, tc.namespace, tc.deletionTimestamp)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	Run(testOpts, api.ClientSet, "")
}

func TestGetClientSet(t *testing.T) {
	restConfig, err := GetConfig("../../mock/kubeconfig", false)
	assert.Nil(t, err)
	assert.NotNil(t, restConfig)

	clientSet, err := GetClientSet(restConfig)
	assert.Nil(t, err)
	assert.NotNil(t, clientSet)

	restConfig, err = GetConfig("../../mock/broken_kubeconfig", false)
	assert.NotNil(t, err)
	assert.Nil(t, restConfig)
}

func TestTerminatePodsWithoutCreating(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)
	var wg sync.WaitGroup

	testOpts := getDefaultOpts()

	wg.Add(1)
	podChannel := make(chan v1.Pod, 10)
	podChannel <- v1.Pod{}
	/*pod, _ := api.createTerminatingPod("demo-pod", "default", nil)
	podChannel <- *pod*/
	go terminatePods(podChannel, &wg, api.ClientSet, logging.GetLogger(), testOpts.GracePeriodSeconds)
	wg.Wait()
}

func TestTerminatePods(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)
	var wg sync.WaitGroup

	testOpts := getDefaultOpts()

	wg.Add(1)
	podChannel := make(chan v1.Pod, 10)
	pod, _ := api.createTerminatingPod("demo-pod", "default", nil)
	podChannel <- *pod
	go terminatePods(podChannel, &wg, api.ClientSet, logging.GetLogger(), testOpts.GracePeriodSeconds)
	wg.Wait()
}
