package main

import (
	"flag"
	"go.uber.org/zap"
	"kube-pod-terminator/pkg/kubernetes"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	masterUrl, kubeConfigPath, namespace string
	tickerIntervalMin, channelCapacity int
	logger *zap.Logger
	err error
)

func init() {
	flag.StringVar(&masterUrl, "masterUrl", "", "(optional) Cluster master ip to access")
	flag.StringVar(&kubeConfigPath, "kubeConfigPath", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"(optional) Kube config file path to access cluster")
	flag.StringVar(&namespace, "namespace", "default", "(optional) Namespace to run on")
	flag.IntVar(&tickerIntervalMin, "tickerIntervalMin", 5, "(optional) This app is a scheduled app, so " +
		"it continuously runs in specified interval")
	flag.IntVar(&channelCapacity, "channelCapacity", 10, "(optional) Channel capacity for concurrency")
	flag.Parse()

	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("fetched arguments", zap.String("masterUrl", masterUrl), zap.String("kubeConfigPath", kubeConfigPath),
		zap.String("namespace", namespace), zap.Int("ticketIntervalMin", tickerIntervalMin), zap.Int("channelCapacity", channelCapacity))
}

func main() {
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	config, err := kubernetes.GetConfig(masterUrl, kubeConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Create an rest client not targeting specific API version
	clientSet, err := kubernetes.GetClientSet(config)
	if err != nil {
		log.Fatalln(err)
	}

	kubernetes.Run(namespace, clientSet, channelCapacity, logger)
	ticker := time.NewTicker(time.Duration(int32(tickerIntervalMin)) * time.Minute)
	for range ticker.C {
		kubernetes.Run(namespace, clientSet, channelCapacity, logger)
	}
	select {}
}