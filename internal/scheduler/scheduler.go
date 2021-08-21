package scheduler

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kube-pod-terminator/internal/logging"
	"log"
	"sync"
)

var logger *zap.Logger

func init() {
	logger = logging.GetLogger()
}

func terminatePod(podChannel chan v1.Pod, wg *sync.WaitGroup, deleteOptions metav1.DeleteOptions,
	clientSet *kubernetes.Clientset, namespace string) {
	for pod := range podChannel {
		err := clientSet.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, deleteOptions)
		if err != nil {
			log.Fatalln(err)
		}
		logger.Info("pod deleted", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace))
		wg.Done()
	}
}

// Run operates the business logic, fetches the terminating pods and terminates them
func Run(namespace string, clientSet *kubernetes.Clientset, channelCapacity int, gracePeriodSeconds int64) {
	pods, err := getTerminatingPods(clientSet, namespace)
	if err != nil {
		logger.Warn("an error occurred while getting terminating pods", zap.Error(err))
	}

	if len(pods) > 0 {
		logger.Info("found pods", zap.Int("podCount", len(pods)), zap.String("namespace", namespace))
		var wg sync.WaitGroup
		deleteOptions := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
		podChannel := make(chan v1.Pod, channelCapacity)
		for i := 0; i < cap(podChannel); i++ {
			go terminatePod(podChannel, &wg, deleteOptions, clientSet, namespace)
		}

		for _, pod := range pods {
			logger.Info("adding pod to podChannel channel", zap.String("name", pod.Name),
				zap.String("namespace", pod.Namespace))
			wg.Add(1)
			podChannel <- pod
		}

		wg.Wait()
		close(podChannel)
	} else {
		logger.Info("no terminating pod found, skipping execution", zap.String("namespace", namespace))
	}
}
