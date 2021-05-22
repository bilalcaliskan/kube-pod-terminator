package main

import (
	"flag"
	"go.uber.org/zap"
	"kube-pod-terminator/pkg/scheduler"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	masterUrl, kubeConfigPath, namespace string
	tickerIntervalMin, channelCapacity   int
	gracePeriodSeconds                   int64
	logger                               *zap.Logger
	err                                  error
	inCluster                            bool
)

func init() {
	flag.StringVar(&masterUrl, "masterUrl", "", "Cluster master ip to access")
	flag.BoolVar(&inCluster, "inCluster", true, "Specify if kube-pod-terminator is running in cluster")
	flag.StringVar(&kubeConfigPath, "kubeConfigPath", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"Kube config file path to access cluster. Required while running out of Kubernetes cluster")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to run on. Defaults to default namespace")
	flag.IntVar(&tickerIntervalMin, "tickerIntervalMin", 5, "Interval of scheduled job to run")
	flag.IntVar(&channelCapacity, "channelCapacity", 10, "Channel capacity for concurrency")
	flag.Int64Var(&gracePeriodSeconds, "gracePeriodSeconds", 30, "Grace period to delete pods")
	flag.Parse()

	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("fetched arguments", zap.String("masterUrl", masterUrl),
		zap.String("kubeConfigPath", kubeConfigPath), zap.String("namespace", namespace),
		zap.Int("ticketIntervalMin", tickerIntervalMin), zap.Int("channelCapacity", channelCapacity),
		zap.Int64("gracePeriodSeconds", gracePeriodSeconds))
}

func main() {
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	config, err := scheduler.GetConfig(masterUrl, kubeConfigPath, inCluster)
	if err != nil {
		log.Fatalln(err)
	}

	// Create an rest client not targeting specific API version
	clientSet, err := scheduler.GetClientSet(config)
	if err != nil {
		log.Fatalln(err)
	}

	scheduler.Run(namespace, clientSet, channelCapacity, gracePeriodSeconds, logger)
	ticker := time.NewTicker(time.Duration(int32(tickerIntervalMin)) * time.Minute)
	for range ticker.C {
		scheduler.Run(namespace, clientSet, channelCapacity, gracePeriodSeconds, logger)
	}
	select {}
}
