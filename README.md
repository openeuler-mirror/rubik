# rubik

## 概述

当前全球云基础设施服务支出费用庞大，然而数据中心用户集群的平均CPU利用率却很低，存在巨大的资源浪费。因此，提升数据中心资源利用率是当前急需解决的一个重要问题。

将多种类型业务混合部署能够显著提升集群资源利用率，但也带来了共峰问题，该问题会导致部分业务服务质量（QoS）受损。如何在提升资源利用率之后，保障业务QoS不受损是技术上的关键挑战。

为此我们提出了Rubik资源利用率提升解决方案，Rubik字面意思为魔方，魔方由Rubik在1974年发明，故Rubik既是人名也指代魔方，在我们的解决方案中，Rubik象征着能够将服务器管理的有条不紊。

Rubik当前支持如下特性：

- [pod CPU优先级配置](./docs/modules.md/#pod-cpu优先级)
- [pod memory优先级配置](./docs/modules.md#pod-内存优先级)
- [pod 访存带宽和LLC限制](./docs/modules.md#dyncache-访存带宽和llc限制)
- [pod blkio配置](./docs/modules.md/#blkio)
- [pod memory cgroup分级](./docs/modules.md/#memory)
- [pod quota burst配置](./docs/modules.md/#quota-burst)

## 编译

拉取源代码：

```sh
git clone https://gitee.com/openeuler/rubik.git
```

进入源码目录编译：

```sh
cd rubik
make
```

制作rubik镜像

```bash
make image
```

将相关部署文件安装到系统中：

```sh
sudo make install
```

## 部署

### 环境准备

- OS: openEuler 21.09/22.03/22.09+
- kubernetes: 1.17.0+

### rubik daemonset部署

在master节点上使用kubectl命令部署rubik daemonset：

```sh
kubectl apply -f /var/lib/rubik/rubik-daemonset.yaml
```

## 常用配置

通过以上方式部署的rubik将以默认配置启动，若用户需要修改rubik的配置，可通过修改rubik-daemonset.yaml文件中的config.json部分后重新部署rubik daemonset。

以下介绍几个常见配置，其他配置详见[配置文档](./docs/config.md)

### Pod优先级自动配置

若在rubik config中配置autoConfig为true开启了Pod自动感知配置功能，用户仅需在部署业务pod时在yaml中通过annotation指定其优先级，部署后rubik会自动感知当前节点pod的创建与更新，并根据用户配置的优先级设置pod优先级。

### 依赖于kubelet的Pod优先级配置

由于自动配置依赖于来自api-server pod创建事件的通知，具有一定的延迟性，无法在进程启动之前及时完成优先级的配置，导致业务性能可能存在抖动。用户可以关闭自动配置，通过修改kubelet，向rubik发送http请求，在更早的时间点调用rubik配置pod优先级，http接口具体使用方法详见[http接口文档](./docs/http_API.md)

### 支持自动校对Pod优先级

rubik支持在启动时对当前节点Pod QoS优先级配置一致性进行校对，这里的一致性是指k8s集群中的配置和rubik对pod优先级的配置之间的一致性。可以通过config选项autoCheck控制是否开启校对功能，默认关闭。若开启校对Pod优先级功能，启动或重启rubik时，rubik会自动校验并更正当前节点pod优先级配置。

## 在离线业务配置示例

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: qosexample
  annotations:
    volcano.sh/preemptable: "true"   # volcano.sh/preemptable为true代表业务为离线业务，false代表业务为在线业务，默认为false
spec:
  containers:
  - name: nginx
    image: nginx
    resources:
      limits:
        memory: "200Mi"
        cpu: "1"
      requests:
        memory: "200Mi"
        cpu: "1"
```

## 注意事项

约束限制详见[约束限制文档](./docs/limitation.md)

## 如何贡献

我们很高兴能有新的贡献者加入！

在一切开始之前，请签署[CLA协议](https://openeuler.org/en/cla.html)

##  版权

rubik遵从**Mulan PSL v2**版权协议
