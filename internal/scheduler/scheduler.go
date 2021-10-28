package scheduler

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"kube-pod-terminator/internal/logging"
	"kube-pod-terminator/internal/options"
	"sync"
)

var (
	logger        *zap.Logger
	opts          *options.KubePodTerminatorOptions
	deleteOptions metav1.DeleteOptions
)

func init() {
	logger = logging.GetLogger()
	opts = options.GetKubePodTerminatorOptions()
	deleteOptions = metav1.DeleteOptions{GracePeriodSeconds: &opts.GracePeriodSeconds}
}

func terminatePods(podChannel chan v1.Pod, wg *sync.WaitGroup, clientSet *kubernetes.Clientset, apiServer string) {
	for pod := range podChannel {
		err := clientSet.CoreV1().Pods(opts.Namespace).Delete(context.Background(), pod.Name, deleteOptions)
		if err != nil {
			logger.Warn("an error occured while deleting pod", zap.String("name", pod.Name),
				zap.String("namespace", pod.Namespace), zap.Bool("inCluster", opts.InCluster),
				zap.String("apiServer", apiServer))
		}
		logger.Info("pod deleted", zap.String("name", pod.Name), zap.String("namespace", pod.Namespace),
			zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))
		wg.Done()
	}
}

func addPodsToChannel(podChannel chan v1.Pod, wg *sync.WaitGroup, podSlice []v1.Pod, state, apiServer string) {
	for _, pod := range podSlice {
		logger.Info("adding pod to podChannel channel", zap.String("state", state),
			zap.String("name", pod.Name), zap.String("namespace", pod.Namespace),
			zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))
		wg.Add(1)
		podChannel <- pod
	}
}

// Run operates the business logic, fetches the terminating and evicted pods and terminates them
func Run(namespace string, clientSet *kubernetes.Clientset, apiServer string) {

	podChannel := make(chan v1.Pod, opts.ChannelCapacity)
	var wg sync.WaitGroup

	terminatingPods, err := getTerminatingPods(clientSet, namespace)
	if err != nil {
		logger.Warn("an error occurred while getting terminating pods, skipping execution", zap.Error(err),
			zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))
		return
	}

	if len(terminatingPods) > 0 {
		logger.Info("found pods", zap.String("state", "terminating"), zap.Int("podCount", len(terminatingPods)),
			zap.String("namespace", namespace), zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))

		addPodsToChannel(podChannel, &wg, terminatingPods, "terminating", apiServer)
	} else {
		logger.Info("no pod found, skipping execution", zap.String("state", "terminating"),
			zap.String("namespace", namespace), zap.Bool("inCluster", opts.InCluster),
			zap.String("apiServer", apiServer))
	}

	if opts.TerminateEvicted {
		evictedPods, err := getEvictedPods(clientSet, namespace)
		if err != nil {
			logger.Warn("an error occurred while getting terminating pods, skipping execution", zap.Error(err),
				zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))
			return
		}

		if len(evictedPods) > 0 {
			logger.Info("found pods", zap.String("state", "evicted"), zap.Int("podCount", len(terminatingPods)),
				zap.String("namespace", namespace), zap.Bool("inCluster", opts.InCluster), zap.String("apiServer", apiServer))

			addPodsToChannel(podChannel, &wg, terminatingPods, "terminating", apiServer)
		} else {
			logger.Info("no pod found, skipping execution", zap.String("state", "evicted"),
				zap.String("namespace", namespace), zap.Bool("inCluster", opts.InCluster),
				zap.String("apiServer", apiServer))
		}
	} else {
		logger.Info("will not terminate evicted pods since --terminateEvicted=false argument passed")
	}

	terminatePods(podChannel, &wg, clientSet, apiServer)
	wg.Wait()
	close(podChannel)

}
