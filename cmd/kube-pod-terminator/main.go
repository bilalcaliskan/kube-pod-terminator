package main

import (
	"context"
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
	kpto              *options.KubePodTerminatorOptions
)

func init() {
	kpto = options.GetKubePodTerminatorOptions()
	logger = logging.GetLogger()
	logger = logger.With(zap.Bool("inCluster", kpto.InCluster))

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

	kubeConfigPathArr = strings.Split(kpto.KubeConfigPaths, ",")
	for _, path := range kubeConfigPathArr {
		go func(p string) {
			logger = logger.With(zap.String("kubeConfigPath", p))
			logger.Info("starting generating clientset for kubeconfig")
			restConfig, err := scheduler.GetConfig(p, kpto.InCluster)
			if err != nil {
				logger.Fatal("fatal error occurred while getting k8s config", zap.String("error", err.Error()))
			}

			clientSet, err := scheduler.GetClientSet(restConfig)
			if err != nil {
				logger.Fatal("fatal error occurred while getting clientset", zap.String("error", err.Error()))
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(kpto.ContextTimeoutSeconds)*time.Second)
			defer cancel()
			scheduler.Run(ctx, kpto.Namespace, clientSet, restConfig.Host)
			ticker := time.NewTicker(time.Duration(kpto.TickerIntervalMinutes) * time.Minute)
			for range ticker.C {
				scheduler.Run(ctx, kpto.Namespace, clientSet, restConfig.Host)
			}
		}(path)
	}

	select {}
}
