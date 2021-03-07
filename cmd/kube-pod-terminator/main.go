package main

import (
	"flag"
	"kube-pod-terminator/pkg/kubernetes"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	masterUrl, kubeConfigPath, namespace *string
	tickerIntervalMin, channelCapacity *int
)

func init() {
	masterUrl = flag.String("masterUrl", "", "(optional) Cluster master ip to access")
	kubeConfigPath = flag.String("kubeConfigPath", filepath.Join(os.Getenv("HOME"), ".kube", "config"),
		"(optional) Kube config file path to access cluster")
	namespace = flag.String("namespace", "default", "(optional) Namespace to run on")
	tickerIntervalMin = flag.Int("tickerIntervalMin", 5, "(optional) This app is a scheduled app, so " +
		"it continuously runs in specified interval")
	channelCapacity = flag.Int("channelCapacity", 10, "(optional) Channel capacity for concurrency")
	flag.Parse()

	log.Printf("Using masterUrl: %s\n", *masterUrl)
	log.Printf("Using kubeConfigPath: %s\n", *kubeConfigPath)
	log.Printf("Using namespace: %s\n", *namespace)
	log.Printf("Using tickerIntervalMin: %d\n", *tickerIntervalMin)
	log.Printf("Using channelCapaciry: %d\n", *channelCapacity)
}

func main() {
	config, err := kubernetes.GetConfig(*masterUrl, *kubeConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Create an rest client not targeting specific API version
	clientSet, err := kubernetes.GetClientSet(config)
	if err != nil {
		log.Fatalln(err)
	}

	kubernetes.Run(*namespace, clientSet, *channelCapacity)
	ticker := time.NewTicker(time.Duration(int32(*tickerIntervalMin)) * time.Minute)
	for _ = range ticker.C {
		kubernetes.Run(*namespace, clientSet, *channelCapacity)
	}
	select {}
}