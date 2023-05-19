# rubik容器调度

当前全球云基础设施服务支出费用庞大，然而数据中心用户集群的平均CPU利用率却很低，存在巨大的资源浪费。因此，提升数据中心资源利用率是当前急需解决的一个重要问题。

将多种类型业务混合部署能够显著提升集群资源利用率，但也带来了共峰问题，该问题会导致部分业务服务质量（QoS）受损。如何在提升资源利用率之后，保障业务QoS不受损是技术上的关键挑战。

为此我们提出了Rubik资源利用率提升解决方案，Rubik字面意思为魔方，魔方由Rubik在1974年发明，故Rubik既是人名也指代魔方，在我们的解决方案中，Rubik象征着能够将服务器管理的有条不紊。

rubik容器调度在业务混合部署的场景下，根据用户对业务的配置（包括QoS分级、cache限制、弹性限流等配置），对资源进行合理调度与隔离，从而实现在保障在线业务的服务体验情况下，提升节点资源利用率。

Rubik当前支持如下[特性列表](./docs/feature.md)：
- [preemption 绝对抢占](./docs/feature.md#preemption-绝对抢占)
- [dynCache 访存带宽和LLC限制](./docs/feature.md#dyncache-访存带宽和llc限制)
- [ioLimit 块设备读写限制](./docs/feature.md#iolimit-块设备读写限制)
- [支持弹性限流](./docs/feature.md#支持弹性限流)
  - [quotaBurst 支持弹性限流内核态解决方案](./docs/feature.md#quotaburst-内核态解决方案)
  - [quotaTurbo 支持弹性限流用户态解决方案](./docs/feature.md#quotaturbo-用户态解决方案)
- [ioCost 支持iocost对IO权重控制](./docs/feature.md#iocost-支持iocost对io权重控制)

## 使用
我们推荐您在kubernetes集群中以daemonset形式执行rubik二进制。

### 运行要求
```
OS: openEuler 21.09/22.03/22.09+
kubernetes: 1.17.0+
```
其余约束详见[约束限制](./docs/limitation.md)

### 通过yaml一键部署rubik
我们提供了一键式脚本供用户使用，并在openEuler容器镜像仓库提供了体验镜像`hub.oepkgs.net/cloudnative/rubik:latest`。

1. 下载rubik yaml脚本
    ```bash
    $ curl -O https://gitee.com/openeuler/rubik/raw/master/hack/rubik-daemonset.yaml
    ```

2. 在master节点上使用kubectl命令部署rubik daemonset：
    ```bash
    $ kubectl apply -f rubik-daemonset.yaml
    ```
    我们在`kube-system`命名空间下创建了名为`rubik-agent-xxx`的pod。
    ```bash
    $ kubectl get pod -A -o wide | grep rubik
    # NAMESPACE     NAME                READY   STATUS    RESTARTS   AGE
    # kube-system   rubik-agent-6bn8n   1/1     Running   0          12m
    ```

### 定制化使用rubik
如果您想自己动手参与修改、编译、按需使用rubik，请参考：
- [如何编译rubik](./docs/usage.md#如何编译rubik)
- [如何使用rubik](./docs/usage.md#如何使用rubik)


如果您在使用过程中遇到任何问题，可先查阅[问题集锦](./docs/trouble.md)。若未能解决您的问题，请直接联系我们或者向社区提出issue。我们热烈欢迎并且感谢您为社区贡献您的力量。


## 如何贡献
我们很高兴能有新的贡献者加入！

在一切开始之前，请签署CLA协议，并且你可能需要了解：
- [新手指南](./docs/getting-started/startup.md)


## 版权
rubik遵从Mulan PSL v2版权协议
