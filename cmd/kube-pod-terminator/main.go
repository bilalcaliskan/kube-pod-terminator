package main

import (
	"github.com/dimiro1/banner"
	"go.uber.org/zap"
	"io/ioutil"
	"kube-pod-terminator/internal/k8s"
	"kube-pod-terminator/internal/logging"
	"kube-pod-terminator/internal/options"
	"os"
	"strings"
	"syscall"
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
	kubeConfigPathArr = strings.Split(kpto.KubeConfigPaths, ",")
	exitSignal := make(chan os.Signal)
	for _, path := range kubeConfigPathArr {
		go func(p string) {
			logger = logger.With(zap.String("kubeConfigPath", p))
			logger.Info("starting generating clientset for kubeconfig")
			restConfig, err := k8s.GetConfig(p, kpto.InCluster)
			if err != nil {
				logger.Fatal("fatal error occurred while getting k8s config", zap.String("error", err.Error()))
			}

			clientSet, err := k8s.GetClientSet(restConfig)
			if err != nil {
				logger.Fatal("fatal error occurred while getting clientset", zap.String("error", err.Error()))
			}

			k8s.Run(kpto.Namespace, clientSet, restConfig.Host)
			if kpto.OneShot {
				exitSignal <- syscall.SIGTERM
				return
			}

			ticker := time.NewTicker(time.Duration(kpto.TickerIntervalMinutes) * time.Minute)
			for range ticker.C {
				k8s.Run(kpto.Namespace, clientSet, restConfig.Host)
			}
		}(path)
	}

	if kpto.OneShot {
		signalCounter := 0
		for range exitSignal {
			signalCounter++
			if signalCounter == len(kubeConfigPathArr) {
				logger.Info("all goroutines sent a SIGTERM, exiting")
				os.Exit(0)
			}
		}
	}

	select {}
}
