package main

import (
	"github.com/dimiro1/banner"
	"go.uber.org/zap"
	"io/ioutil"
	"kube-pod-terminator/pkg/logging"
	"kube-pod-terminator/pkg/options"
	"kube-pod-terminator/pkg/scheduler"
	"os"
	"strings"
	"time"
)

var logger *zap.Logger

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
	config, err := scheduler.GetConfig(kpto.MasterUrl, kpto.KubeConfigPath, kpto.InCluster)
	if err != nil {
		logger.Fatal("a fatal error occured while getting config", zap.Error(err))
	}

	// Create an rest client not targeting specific API version
	clientSet, err := scheduler.GetClientSet(config)
	if err != nil {
		logger.Fatal("a fatal error occured while getting clientset", zap.Error(err))
	}

	scheduler.Run(kpto.Namespace, clientSet, kpto.ChannelCapacity, kpto.GracePeriodSeconds)
	ticker := time.NewTicker(time.Duration(int32(kpto.TickerIntervalMin)) * time.Minute)
	for range ticker.C {
		scheduler.Run(kpto.Namespace, clientSet, kpto.ChannelCapacity, kpto.GracePeriodSeconds)
	}
	select {}
}
