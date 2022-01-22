package k8s

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

type FakeAPI struct {
	ClientSet kubernetes.Interface
	Namespace string
}

func (fAPI *FakeAPI) createPod(name string) (*v1.Pod, error) {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         fAPI.Namespace,
			DeletionTimestamp: &metav1.Time{Time: time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC)},
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

	pod, err := fAPI.ClientSet.CoreV1().Pods(fAPI.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
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

func TestRun(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)

	cases := []struct {
		caseName, podName, namespace string
	}{
		{
			caseName: "case1",
			podName:  "varnish-pod-1",
		},
		{
			caseName: "case2",
			podName:  "varnish-pod-2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			pod, err := api.createPod(tc.podName)
			assert.Nil(t, err)
			assert.NotNil(t, pod)
			time.Sleep(2 * time.Second)
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Run(ctx, api.Namespace, api.ClientSet, "")
}

func TestGetClientSet(t *testing.T) {
	restConfig, err := GetConfig("../../mock/kubeconfig", false)
	assert.Nil(t, err)
	assert.NotNil(t, restConfig)

	clientSet, err := GetClientSet(restConfig)
	assert.Nil(t, err)
	assert.NotNil(t, clientSet)
}

/*func TestTerminatePods(t *testing.T) {
	api := getFakeAPI()
	assert.NotNil(t, api)
	var wg sync.WaitGroup
	podChannel := make(chan v1.Pod, opts.ChannelCapacity)
	podChannel <- v1.Pod{}
	terminatePods(podChannel, &wg, api.ClientSet, "")
}*/
