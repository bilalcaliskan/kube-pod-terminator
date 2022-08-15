package k8s

import (
	"context"
	"log"
	"sync"

	"github.com/bilalcaliskan/kube-pod-terminator/internal/logging"
	"github.com/bilalcaliskan/kube-pod-terminator/internal/options"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// terminatePods does the real job, terminates the items in the v1.Pod channel with specified clientSet
func terminatePods(podChannel chan v1.Pod, wg *sync.WaitGroup, clientSet kubernetes.Interface, logger *zap.Logger, gracePeriodSeconds int64) {
	for pod := range podChannel {
		if err := clientSet.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name,
			metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
			logger.Warn("an error occured while deleting pod", zap.String("name", pod.Name),
				zap.String("error", err.Error()))
			wg.Done()
			continue
		}

		logger.Info("pod successfully terminated", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace))
		wg.Done()
	}
}

// addPodsToChannel adds items of v1.Pod slice to specified v1.Pod channel
func addPodsToChannel(podChannel chan v1.Pod, wg *sync.WaitGroup, podSlice []v1.Pod, state string, logger *zap.Logger) {
	for _, pod := range podSlice {
		logger.Info("adding pod to podChannel channel", zap.String("name", pod.Name),
			zap.String("namespace", pod.Namespace), zap.String("state", state))
		podChannel <- pod
		wg.Add(1)
	}
}

// Run operates the business logic, fetches the terminating and evicted pods and terminates them
func Run(opts *options.KubePodTerminatorOptions, clientSet kubernetes.Interface, apiServer string) {
	logger := logging.GetLogger().With(zap.String("apiServer", apiServer))
	podChannel := make(chan v1.Pod, 50)
	var wg sync.WaitGroup

	terminatingPods, err := getTerminatingPods(clientSet, opts.Namespace, opts.TerminatingStateMinutes)
	if err != nil {
		logger.Warn("an error occurred while getting terminating pods, skipping execution", zap.Error(err))
		return
	}

	if len(terminatingPods) > 0 {
		logger.Info("found pods", zap.String("state", "terminating"), zap.Int("podCount", len(terminatingPods)))
		log.Println(cap(podChannel))
		log.Println(len(podChannel))
		addPodsToChannel(podChannel, &wg, terminatingPods, "terminating", logger)
	} else {
		logger.Info("no pod found, skipping execution", zap.String("state", "terminating"))
	}

	if opts.TerminateEvicted {
		evictedPods, err := getEvictedPods(clientSet, opts.Namespace)
		if err != nil {
			logger.Warn("an error occurred while getting evicted pods, skipping execution", zap.Error(err))
			return
		}

		if len(evictedPods) > 0 {
			logger.Info("found pods", zap.String("state", "evicted"), zap.Int("podCount", len(evictedPods)))
			addPodsToChannel(podChannel, &wg, evictedPods, "evicted", logger)
		} else {
			logger.Info("no pod found, skipping execution", zap.String("state", "evicted"))
		}
	} else {
		logger.Info("will not terminate evicted pods since --terminateEvicted=false argument passed")
	}

	close(podChannel)

	go terminatePods(podChannel, &wg, clientSet, logger, opts.GracePeriodSeconds)
	wg.Wait()
}
