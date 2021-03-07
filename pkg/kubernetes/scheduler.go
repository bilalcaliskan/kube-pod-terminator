package kubernetes

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"sync"
)

func forceKillPod(podChannel chan v1.Pod, wg *sync.WaitGroup, deleteOptions metav1.DeleteOptions,
	clientSet *kubernetes.Clientset, namespace string) {
	for pod := range podChannel {
		err := clientSet.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, deleteOptions)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Pod %s force deleted!\n", pod.Name)
		wg.Done()
	}
}

func Run(namespace string, clientSet *kubernetes.Clientset, channelCapacity int) {
	pods, err := getTerminatingPods(clientSet, namespace)

	if err != nil {
		log.Fatalln(err)
	}

	if len(pods) > 0 {
		log.Printf("%d pods found on namespace %s, starting execution!\n", len(pods), namespace)
		var wg sync.WaitGroup
		var deleteOptions metav1.DeleteOptions
		var zero int64 = 0
		deleteOptions = metav1.DeleteOptions{GracePeriodSeconds: &zero}
		podChannel := make(chan v1.Pod, channelCapacity)
		for i := 0; i < cap(podChannel); i++ {
			go forceKillPod(podChannel, &wg, deleteOptions, clientSet, namespace)
		}

		for i, pod := range pods {
			wg.Add(1)
			log.Printf("Adding pod '%d - %s' to the podChannel channel!\n", i, pod.Name)
			podChannel <- pod
		}

		wg.Wait()
		close(podChannel)
	} else {
		log.Printf("No terminating pod found on namespace %s, so skipping execution...\n", namespace)
	}
}