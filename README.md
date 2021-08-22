# Kube Pod Terminator
[![CI](https://github.com/bilalcaliskan/kube-pod-terminator/workflows/CI/badge.svg?event=push)](https://github.com/bilalcaliskan/kube-pod-terminator/actions?query=workflow%3ACI)
[![Docker pulls](https://img.shields.io/docker/pulls/bilalcaliskan/kube-pod-terminator)](https://hub.docker.com/r/bilalcaliskan/kube-pod-terminator/)
[![Go Report Card](https://goreportcard.com/badge/github.com/bilalcaliskan/kube-pod-terminator)](https://goreportcard.com/report/github.com/bilalcaliskan/kube-pod-terminator)

On some Kubernetes versions, there is a problem that pods stuck in `Terminating` status on some circumstances. This tool runs
in Kubernetes cluster and connects to the kube-apiserver, discovers Terminating pods which are in `Terminating` status
more than 30 minutes.

## Deployment
Kube-pod-terminator can be deployed as Kubernetes deployment or standalone installation. You can use [sample config file](config/sample.yaml) to deploy your Kubernetes cluster.
This file also creates required `Role` and `RoleBindings` to take actions on problematic pods.

```shell
$ kubectl create -f config/sample.yaml
```

### Customization
Kube-pod-terminator can be customized with several command line arguments. You can pass arguments
via [sample config file](config/sample.yaml). Here is the list of arguments you can pass:

```
--inCluster             Specify if kube-pod-terminator is running in cluster. Defaults to true
--masterUrl             Cluster master ip to access. Defaults to "".
--kubeConfigPath        Kube config file path to access cluster. Required while running out of Kubernetes cluster.
--namespace             Namespace to run on. Defaults to "default" namespace.
--tickerIntervalMin     Kube-pod-terminator runs as scheduled job. This argument is the interval of scheduled job to run. Defaults to 5.
--channelCapacity       Channel capacity for concurrency. Defaults to 10.
--gracePeriodSeconds    Grace period to delete pods. Defaults to 30.
```

### Development
This project requires below tools while developing:
- [pre-commit](https://pre-commit.com/)
- [golangci-lint](https://golangci-lint.run/usage/install/) - required by [pre-commit](https://pre-commit.com/)

### How kube-pod-terminator handles authentication/authorization with kube-apiserver?

kube-pod-terminator uses [client-go](https://github.com/kubernetes/client-go) to interact
with `kube-apiserver`. [client-go](https://github.com/kubernetes/client-go) uses the [service account token](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
mounted inside the Pod at the `/var/run/secrets/kubernetes.io/serviceaccount` path while initializing the client.

If you have RBAC enabled on your cluster, when you applied the sample deployment file [config/sample.yaml](config/sample.yaml),
it will create required serviceaccount, role and rolebinding and then use that serviceaccount to be used
by our kube-pod-terminator pods.

If RBAC is not enabled on your cluster, please follow [that documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to enable it.
