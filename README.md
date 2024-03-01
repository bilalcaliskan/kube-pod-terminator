# Kube Pod Terminator
[![CI](https://github.com/bilalcaliskan/kube-pod-terminator/workflows/CI/badge.svg?event=push)](https://github.com/bilalcaliskan/kube-pod-terminator/actions?query=workflow%3ACI)
[![Docker pulls](https://img.shields.io/docker/pulls/bilalcaliskan/kube-pod-terminator)](https://hub.docker.com/r/bilalcaliskan/kube-pod-terminator/)
[![Go Report Card](https://goreportcard.com/badge/github.com/bilalcaliskan/kube-pod-terminator)](https://goreportcard.com/report/github.com/bilalcaliskan/kube-pod-terminator)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=bilalcaliskan_kube-pod-terminator&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=bilalcaliskan_kube-pod-terminator)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=bilalcaliskan_kube-pod-terminator&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=bilalcaliskan_kube-pod-terminator)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=bilalcaliskan_kube-pod-terminator&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=bilalcaliskan_kube-pod-terminator)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=bilalcaliskan_kube-pod-terminator&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=bilalcaliskan_kube-pod-terminator)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=bilalcaliskan_kube-pod-terminator&metric=coverage)](https://sonarcloud.io/summary/new_code?id=bilalcaliskan_kube-pod-terminator)
[![Release](https://img.shields.io/github/release/bilalcaliskan/kube-pod-terminator.svg)](https://github.com/bilalcaliskan/kube-pod-terminator/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/bilalcaliskan/kube-pod-terminator)](https://github.com/bilalcaliskan/kube-pod-terminator)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

On some Kubernetes versions, there is a problem that pods stuck in **Terminating** state on some circumstances. This tool
connects to the **kube-apiserver**, discovers Terminating pods which are in **Terminating** status more than **--terminatingStateMinutes**
minutes, which is defaults to 30 minutes.

This tool also discovers pods which are at **Evicted** state if **--terminate-evicted** flag passed(enabled by default) and
clears them all.

Please note that **kube-pod-terminator** can work in below modes:
- Outside of Kubernetes cluster as a CLI (**--one-shot** should be passed, default behavior)
- Inside Kubernetes cluster as Deployment (**--in-cluster=true** should be passed)
- Outside of Kubernetes cluster as binary (**--in-cluster=false --one-shot=false** should be passed)

Please refer to [Installation section](#installation) for more information.

## Notable Features
- All namespaces support
- Multi clusters support
- Ability to run in Kubernetes as Deployment
- Ability to run outside of Kubernetes as binary (Linux, Darwin)
- Ability to run outside of Kubernetes as CLI (oneshot app, not scheduled)
- Homebrew

## Configuration
Kube-pod-terminator can be customized with several command line arguments. You can pass arguments
via [sample deployment file](deployments/sample_single_namespace.yaml) or directly to the binary. Here is the list of arguments you can pass:

```
Usage:
  kube-pod-terminator [flags]

Flags:
      --grace-period-seconds int          grace period to delete target pods (default 30)
  -h, --help                              help for kube-pod-terminator
      --in-cluster                        specify if kube-pod-terminator is running in cluster
      --kubeconfig-paths string          comma separated list of kubeconfig file paths to access with the cluster (default "/home/joshsagredo/.kube/config")
      --namespace string                  target namespace to run on (default "all")
      --one-shot                          specifier to run kube-pod-terminator only one time instead of continuously running in the background. should be true if you are using app as CLI. (default true)
      --terminate-evicted                 terminate evicted pods in specified namespaces (default true)
      --terminating-state-minutes int32   terminate stucked pods in terminating state which are more than that value (default 30)
      --ticker-interval-minutes int32     interval of scheduled job to run (default 5)
  -v, --verbose                           verbose output of the logging library (default false)
      --version                           version for kube-pod-terminator
```

## Installation
Kube-pod-terminator can be deployed as Kubernetes deployment or standalone installation

### Kubernetes
You can use [sample deployment file](deployments/sample_single_namespace.yaml) to deploy your Kubernetes cluster.
This file also creates required **Role** and **RoleBindings** to take actions on problematic pods.

```shell
$ kubectl create -f deployments/sample_single_namespace.yaml
```

### All namespaces support
By default, kube-pod-terminator runs to terminate pods in `default` namespace. But that behavior can be changed with
`namespace` flag. You can see the example Kubernetes manifest file [deployment/sample_all_namespaces.yaml](deployments/sample_all_namespaces.yaml).
Keep in mind that this file creates necessary `ClusterRole` and `ClusterRoleBinding` to be able to take proper actions on all
namespaces.
```
--namespace=all
```

### Multi Cluster support
kube-pod-terminator can terminate the pods of multiple clusters if multiple kubeconfig file path is provided
to **--kubeconfig-paths** flag.

If you run the kube-pod-terminator inside a Kubernetes cluster, it manages the terminating pods of that current
cluster by default. But keep in mind that if you want to manage terminating pods on multiple clusters
and run kube-pod-terminator inside a Kubernetes cluster, you should mount multiple kubeconfig files as configmap or secret
into pod and pass below arguments in your Deployment config:
```
--in-cluster=false
--kubeconfig-paths=/tmp/kubeconfig1,/tmp/kubeconfig2,/tmp/kubeconfig3
```

You can check [deployment/sample_external_clusters.yaml](deployments/sample_external_clusters.yaml) as example.

But before creating deployment file, you should create configmaps from your desired kubeconfigs like below:
```shell
$ kubectl create configmap cluster1-config --from-file=${YOUR_CLUSTER1_CONFIG_PATH}
$ kubectl create configmap cluster2-config --from-file=${YOUR_CLUSTER2_CONFIG_PATH}
$ kubectl create configmap cluster3-config --from-file=${YOUR_CLUSTER3_CONFIG_PATH}
```

### Homebrew
This project can be installed with [Homebrew](https://brew.sh/):
```
brew tap bilalcaliskan/tap
brew install bilalcaliskan/tap/kube-pod-terminator
```

### Binary
Binary can be downloaded from [Releases](https://github.com/bilalcaliskan/kube-pod-terminator/releases) page. You can
use that method to run kube-pod-terminator outside of a Kubernetes cluster.

After then, you can simply run binary by providing required command line arguments:
```shell
$ ./kube-pod-terminator --in-cluster=false --kubeconfig-paths ~/.kube/config
```

> Critical command line arguments while running kube-pod-terminator as standalone application are **--inCluster**, **--kubeConfigPaths**

## Development
This project requires below tools while developing:
- [Golang 1.21](https://golang.org/doc/go1.21)
- [pre-commit](https://pre-commit.com/)
- [golangci-lint](https://golangci-lint.run/usage/install/) - required by [pre-commit](https://pre-commit.com/)
- [gocyclo](https://github.com/fzipp/gocyclo) - required by [pre-commit](https://pre-commit.com/)

After you installed [pre-commit](https://pre-commit.com/), simply run below command to prepare your development environment:
```shell
$ pre-commit install -c build/ci/.pre-commit-config.yaml
```

## License
Apache License 2.0
