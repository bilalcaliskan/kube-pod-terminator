package main

import (
	"github.com/dimiro1/banner"
	"go.uber.org/zap"
	"io/ioutil"
	"kube-pod-terminator/internal/logging"
	"kube-pod-terminator/internal/options"
	"kube-pod-terminator/internal/scheduler"
	"os"
	"strings"
	"time"
)

var (
	logger            *zap.Logger
	kubeConfigPathArr []string
)

func init() {
	logger = logging.GetLogger()

	bannerBytes, _ := ioutil.ReadFile("banner.txt")
	banner.Init(os.Stdout, true, false, strings.NewReader(string(bannerBytes)))
}

func main() {
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	kpto := options.GetKubePodTerminatorOptions()
	kubeConfigPathArr = strings.Split(kpto.KubeConfigPaths, ",")
	for _, path := range kubeConfigPathArr {
		go func(p string) {
			logger.Info("starting generating clientset for kubeconfig", zap.Bool("inCluster", kpto.InCluster),
				zap.String("kubeConfigPath", p))
			restConfig, err := scheduler.GetConfig(p, kpto.InCluster)
			if err != nil {
				logger.Fatal("fatal error occurred while getting k8s config", zap.String("error", err.Error()),
					zap.Bool("inCluster", kpto.InCluster), zap.String("kubeConfigPath", p))
			}

			clientSet, err := scheduler.GetClientSet(restConfig)
			if err != nil {
				logger.Fatal("fatal error occurred while getting clientset", zap.String("error", err.Error()),
					zap.Bool("inCluster", kpto.InCluster), zap.String("kubeConfigPath", p))
			}

			scheduler.Run(kpto.Namespace, clientSet, restConfig.Host)
			ticker := time.NewTicker(time.Duration(kpto.TickerIntervalMin) * time.Minute)
			for range ticker.C {
				scheduler.Run(kpto.Namespace, clientSet, restConfig.Host)
			}
		}(path)
	}

	select {}
}
