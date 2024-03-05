# Rubik: Hybrid-Deployment Engine for Containers

Cloud infrastructure is expensive. However, the average CPU utilization in clusters is low, causing huge resource waste. Therefore, improving the resource utilization in a data center is quite important.

Hybrid deployment of multiple types of services can significantly improve cluster resource utilization, but also bring the problem of co-peak hours, which may impair the quality of service (QoS) of some services. How to ensure that service QoS is not affected after resource utilization is improved is a key technical challenge.

To solve this problem, we proposed the Rubik resource utilization improvement solution. Rubik literally means Rubik's Cube. Rubik's Cube was invented by Rubik in 1974. Therefore, Rubik is both a person and a Rubik's Cube. In our solution, Rubik symbolizes the ability to manage servers in an orderly manner.

When services are deployed in hybrid mode, the rubik engine helps to improve CPU utilization based on user configurations (including QoS tiering, cache limit, and elastic rate limiting), properly schedules and isolates resources to improve node resource utilization while ensuring online service experience.

Rubik currently supports the following [features](./docs/feature.md) :

 *  [Absolute-preemptive CPU Scheduling and MEM allocation](./docs/feature.md#preemption-绝对抢占)
 *  [dynCache: memory bandwidth and LLC limit](./docs/feature.md#dyncache-内存带宽和LLC限制)
 *  [ioLimit: Block device read/write limit](./docs/feature.md#iolimit-块设备读写限制)
 *  [Elastic CPU limiting](./docs/feature.md#支持弹性限流)
	 * [quotaBurst: kernel-space solution](./docs/feature.md#quotaburst-内核态解决方案)
	 * [quotaTurbo: user-space solution](./docs/feature.md#quotaturbo-用户态解决方案)
 *  [iocost: I/O weight control](./docs/feature.md#iocost-支持iocost对io权重控制)

## How to use

We recommend that you execute the rubik as daemonset in the Kubernetes cluster.

### OS Requirements

```
OS: openEuler 21.09/22.03/22.09+

kubernetes: 1.17.0+
```

For other restrictions, see [Constraints](./docs/limitation.md).


### Deployment

We provide one-click script for quick experience, and the required rubik image is in `hub.oepkgs.net/cloudnative/rubik:latest`.

1.  Download the rubik yaml file to master node in a kubernetes cluster.

```bash
$ curl -O https://gitee.com/openeuler/rubik/raw/master/hack/rubik-daemonset.yaml
```

2.  Run the kubectl command to deploy rubik daemonset.

```bash
$ kubectl apply -f rubik-daemonset.yaml
```

Then a pod named `rubik-agent-xxx` is running under the `kube-system` namespace.

```bash
$ kubectl get pod -A -o wide | grep rubik
# NAMESPACE     NAME                READY   STATUS    RESTARTS   AGE
# kube-system   rubik-agent-6bn8n   1/1     Running   0          12m
```

### Customized 

If you want to modify, compile, and use rubik on demand, please refer to:

 *  [How to compile rubik](./docs/usage.md#如何编译rubik)
 *  [How to use rubik](./docs/usage.md#如何使用rubik)	

If you encounter any problem, please refer to [troubles](./docs/trouble.md). If it is not resolved, please raise an issue. We warmly welcome and thank you for contributing to the community.

## How to contribute

We are happy to provide guidance for the new contributors.

Please sign the [CLA](https://openeuler.org/en/cla.html) before contributing.

And you may need to read [Beginner's Guide](./docs/getting-started/startup.md)	

## Licensing

Rubik is licensed under the Mulan PSL v2.

