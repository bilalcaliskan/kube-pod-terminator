# Kube Pod Terminator
[![CI](https://github.com/bilalcaliskan/kube-pod-terminator/workflows/CI/badge.svg?event=push)](https://github.com/bilalcaliskan/kube-pod-terminator/actions?query=workflow%3ACI)
[![Docker pulls](https://img.shields.io/docker/pulls/bilalcaliskan/kube-pod-terminator)](https://hub.docker.com/r/bilalcaliskan/kube-pod-terminator/)
[![Go Report Card](https://goreportcard.com/badge/github.com/bilalcaliskan/kube-pod-terminator)](https://goreportcard.com/report/github.com/bilalcaliskan/kube-pod-terminator)
[![codecov](https://codecov.io/gh/bilalcaliskan/kube-pod-terminator/branch/master/graph/badge.svg)](https://codecov.io/gh/bilalcaliskan/kube-pod-terminator)
[![Release](https://img.shields.io/github/release/bilalcaliskan/kube-pod-terminator.svg)](https://github.com/bilalcaliskan/kube-pod-terminator/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/bilalcaliskan/kube-pod-terminator)](https://github.com/bilalcaliskan/kube-pod-terminator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

On some Kubernetes versions, there is a problem that pods stuck in **Terminating** status on some circumstances. This tool runs
in Kubernetes cluster and connects to the **kube-apiserver**, discovers Terminating pods which are in **Terminating** status
more than 30 minutes.

## Configuration
Kube-pod-terminator can be customized with several command line arguments. You can pass arguments
via [sample deployment file](deployment/sample.yaml) or directly to the binary. Here is the list of arguments you can pass:

```
--inCluster             Specify if kube-pod-terminator is running in cluster. Defaults to true
--kubeConfigPath        Comma seperated list of kubeconfig files path to access clusters. Required while running out of Kubernetes cluster.
--namespace             Namespace to run on. Defaults to "default" namespace.
--tickerIntervalMin     Kube-pod-terminator runs as scheduled job. This argument is the interval of scheduled job to run. Defaults to 5.
--channelCapacity       Channel capacity for concurrency. Defaults to 10.
--gracePeriodSeconds    Grace period to delete pods. Defaults to 30.
```

### Multi Cluster support
kube-pod-terminator can terminate the pods of multiple clusters if multiple kubeconfig file path is provided
to **--kubeConfigPath** flag.

If you run the kube-pod-terminator inside a Kubernetes cluster, it manages the terminating pods of that current
cluster by default. But keep in mind that if you want to manage terminating pods on multiple clusters
and run kube-pod-terminator inside a Kubernetes cluster, you should mount multiple kubeconfig files as configmap or secret
into pod and pass below arguments in your Deployment config:
```
--inCluster=false
--kubeConfigPath=/tmp/kubeconfig1,/tmp/kubeconfig2,/tmp/kubeconfig3
```

You can check [deployment/sample_external_clusters.yaml](deployment/sample_external_clusters.yaml) as example.

But before creating deployment file, you should create configmaps from your desired kubeconfigs like below:
```shell
$ kubectl create configmap cluster1-config --from-file=${YOUR_CLUSTER1_CONFIG_PATH}
$ kubectl create configmap cluster2-config --from-file=${YOUR_CLUSTER2_CONFIG_PATH}
$ kubectl create configmap cluster3-config --from-file=${YOUR_CLUSTER3_CONFIG_PATH}
```

## Installation
Kube-pod-terminator can be deployed as Kubernetes deployment or standalone installation

### Kubernetes
You can use [sample deployment file](deployment/sample.yaml) to deploy your Kubernetes cluster.
This file also creates required **Role** and **RoleBindings** to take actions on problematic pods.

```shell
$ kubectl create -f deployment/sample.yaml
```

### Binary
Binary can be downloaded from [Releases](https://github.com/bilalcaliskan/kube-pod-terminator/releases) page. You can
use that method to run kube-pod-terminator outside of a Kubernetes cluster.

After then, you can simply run binary by providing required command line arguments:
```shell
$ ./kube-pod-terminator --inCluster false --kubeConfigPath ~/.kube/config
```

> Critical command line arguments while running kube-pod-terminator as standalone application are **--inCluster**, **--kubeConfigPaths**

## Development
This project requires below tools while developing:
- [Golang 1.16](https://golang.org/doc/go1.16)
- [pre-commit](https://pre-commit.com/)
- [golangci-lint](https://golangci-lint.run/usage/install/) - required by [pre-commit](https://pre-commit.com/)

## License
Apache License 2.0

## How kube-pod-terminator handles authentication/authorization with kube-apiserver?

kube-pod-terminator uses [client-go](https://github.com/kubernetes/client-go) to interact
with `kube-apiserver`. [client-go](https://github.com/kubernetes/client-go) uses the [service account token](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
mounted inside the Pod at the `/var/run/secrets/kubernetes.io/serviceaccount` path while initializing the client.

If you have RBAC enabled on your cluster, when you applied the sample deployment file [deployment/sample.yaml](deployment/sample.yaml),
it will create required serviceaccount, role and rolebinding and then use that serviceaccount to be used
by our kube-pod-terminator pods.

If RBAC is not enabled on your cluster, please follow [that documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to enable it.
