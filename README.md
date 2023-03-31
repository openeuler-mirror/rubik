# rubik容器调度

当前全球云基础设施服务支出费用庞大，然而数据中心用户集群的平均CPU利用率却很低，存在巨大的资源浪费。因此，提升数据中心资源利用率是当前急需解决的一个重要问题。

将多种类型业务混合部署能够显著提升集群资源利用率，但也带来了共峰问题，该问题会导致部分业务服务质量（QoS）受损。如何在提升资源利用率之后，保障业务QoS不受损是技术上的关键挑战。

为此我们提出了Rubik资源利用率提升解决方案，Rubik字面意思为魔方，魔方由Rubik在1974年发明，故Rubik既是人名也指代魔方，在我们的解决方案中，Rubik象征着能够将服务器管理的有条不紊。

rubik容器调度在业务混合部署的场景下，根据用户对业务的配置（包括QoS分级、cache限制、弹性限流等配置），对资源进行合理调度与隔离，从而实现在保障在线业务的服务体验情况下，提升节点资源利用率。

Rubik当前支持如下[特性列表](https://gitee.com/openeuler/rubik/blob/master/docs/feature.md)：
- preemption 绝对抢占
- dynCache 访存带宽和LLC限制
- ioLimit 块设备读写限制
- dynMemory 内存分级回收
- 支持弹性限流
  - quotaBurst 支持弹性限流内核态解决方案
  - quotaTurbo 支持弹性限流用户态解决方案
- ioCost 支持iocost对IO权重控制

## 编译
编译详见[如何编译rubik](https://gitee.com/openeuler/rubik/blob/master/docs/usage.md#如何编译rubik)
### 依赖
```
golang >= 1.13
```
### 步骤
1. 拉取源代码：
```bash
git clone https://gitee.com/openeuler/rubik.git
```

2. 进入源码目录编译：
```bash
cd rubik
make
# 制作rubik镜像
make image
```

3. 将相关部署文件安装到系统中：
```bash
sudo make install
```
## 使用

使用详见[如何使用rubik](https://gitee.com/openeuler/rubik/blob/master/docs/usage.md#如何使用rubik)

### 环境准备
```
OS: openEuler 21.09/22.03/22.09+
kubernetes: 1.17.0+
```
### 运行
在master节点上使用kubectl命令部署rubik daemonset：
```bash
kubectl apply -f /var/lib/rubik/rubik-daemonset.yaml
```


## 如何贡献
我们很高兴能有新的贡献者加入！

在一切开始之前，请签署CLA协议

## 版权
rubik遵从Mulan PSL v2版权协议
