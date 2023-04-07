# 特性介绍
在rubik中，每一个特性以服务形式运行在后台。rubik根据用户配置（config.json）按需启动对应服务。下文是对各个服务的介绍。

## preemption 绝对抢占
rubik支持业务优先级配置，针对在离线业务混合部署的场景，确保在线业务相对离线业务的资源抢占。目前仅支持CPU资源和内存资源。使用该特性，用户需要手动为业务指定业务类型，即在业务pod的yaml文件中增加注解`volcano.sh/preemptable`。业务优先级配置示例如下：

```yaml
annotations:
    volcano.sh/preemptable: true
```
> 在rubik中，所有特性均通过识别该注解作为业务在离线标志。
> true代表业务为离线业务。
> false代表业务为在线业务。

### CPU绝对抢占
针对在离线业务混合部署的场景，确保在线业务相对离线业务的CPU资源抢占。
#### 前置条件
-   内核支持针对cgroup的cpu优先级配置，cpu子系统存在接口`cpu.qos_level`。建议使用内核版本openEuler-22.03+。

### 内存绝对抢占
针对在离线业务混合部署的场景，确保OOM时优先kill离线业务。

#### 前置条件
-   内核支持针对cgroup的memory优先级配置，memory子系统存在接口`memory.qos_level`。建议使用内核版本openEuler-22.03+。
-   开启内存优先级支持: `echo 1 > /proc/sys/vm/memcg_qos_enable`

## dynCache 访存带宽和LLC限制
rubik支持业务的Pod访存带宽(memory bandwidth)和LLC(Last Level Cache)限制，通过限制离线业务的访存带宽/LLC使用，减少其对在线业务的干扰。

该特性依赖于物理机支持intel RDT（x86）和mapm（arm）功能。其将集群中的业务划分为5个控制组(分别为`rubik_max`、`rubik_high`、`rubik_middle`、`rubik_low`、`rubik_dynamic`)，每一个控制组会根据配置限制业务对访存待宽和最后一级缓存的使用。rubik启动后，将水位线写入对应控制组的schemata。其中，`rubik_high`、`rubik_middle`、`rubik_low`控制组对应的水位线是全局的，可在`dynCache`中配置；max控制组为默认最大值, dynamic控制组初始水位线和low控制组一致。

rubik支持两种方式为业务Pod配置访存带宽和LLC控制组：
- 全局方式
用户可在rubik的全局参数中配置`defaultLimitMode`字段，rubik会自动为离线业务Pod（即绝对抢占特性中的注解`volcano.sh/preemptable`）配置控制组。
  - 取值为`static`时，pod将被加入到`rubik_max`控制组。
  - 取值为`dynamic`时，pod将被加入到`rubik_dynamic`控制组。
- 手动指定
用户可手动通过为业务Pod增加注解`volcano.sh/cache-limit`设置其控制组（我们建议不建议将）, 如下列配置的pod将被加入rubik_low控制组:
```yaml
annotations:
    volcano.sh/cache-limit: "low"
```
**rubik dynamic控制组**：
当Pod控制组为`rubik_dynamic`时，rubik通过采集当前节点在线业务pod的`cache miss`和`llc miss`指标，动态调整`rubik_dynamic`控制组的水位线，实现对dynamic控制组内pod的动态控制。

#### 约束限制
- cache/访存限制功能仅支持物理机，不支持虚拟机。
    -   X86物理机，需要OS支持且开启intel RDT的CAT和MBA功能，内核启动项cmdline需要添加`rdt=l3cat,mba`
    -   ARM物理机，需要OS支持且开启mpam功能，内核启动项需要添加`mpam=acpi`。
- 由于内核限制，RDT mode当前不支持pseudo-locksetup模式。
- rubik需挂载主机目录 `/sys/fs/resctrl`，且运行过程中不可卸载。
- rubik需要权限: `SYS_ADMIN`才能正确设置`/sys/fs/resctrl`下的文件。
- rubik与主机共享`pid namespace`，以便获取业务容器进程在主机上的pid。
- 非手动指定，`dynCache`仅针对离线业务，对在线业务不生效。
- 若业务容器运行过程中被手动重启（容器ID不变但容器进程PID变化），针对该容器的dynCache无法生效。
- 业务容器启动并已设置dynCache级别后，不支持对其限制级别进行修改。
- 动态限制组的调控灵敏度受到rubik配置文件内adjustInterval、perfDuration值以及节点在线业务pod数量的影响，每次调整（若干扰检测结果为需要调整）间隔在区间`[adjustInterval+perfDuration, adjustInterval+perfDuration*pod数量]`内波动，用户可根据灵敏度需求调整配置项。

## ioLimit 块设备读写限制
通过cgroup提供的blkio能力限制pod对io资源的使用。
用户需手动为业务Pod配置`volcano.sh/blkio-limit`注解，其格式如下：
```yaml
volcano.sh/blkio-limit: '{"device_read_bps":[{"device":"/dev/sda1","value":"10485760"}, {"device":"/dev/sda","value":"20971520"}],
                "device_write_bps":[{"device":"/dev/sda1","value":"20971520"}],
                "device_read_iops":[{"device":"/dev/sda1","value":"200"}],
                "device_write_iops":[{"device":"/dev/sda1","value":"300"}]}'
```

配置内容:
| 配置键           | 描述                                                                                                                                  |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| device_read_bps  | 用于设定设备执行“读”操作字节的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备，value限定上限值，单位为byte   |
| device_read_iops | 用于设定设备执行“读”操作次数的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备                                |
| device_write_bps | 用于设定设备执行 “写” 操作次数的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备，value限定上限值，单位为byte |
| device_write_iops| 用于设定设备执行“写”操作字节的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备 |

> 说明：
> 1. 配置的key:value和cgroup的key:value的配置规则一致:
> 2. 写入时会转换成环境page size的倍数
> 3. 只有minor为0的device配置才会生效
> 4. 如果取消限速，可将值设为0

## 支持弹性限流
为有效解决由业务CPU限流导致QoS下降的问题，rubik容器提供了弹性限流功能，允许容器使用额外的CPU资源，从而保证业务的平稳运行。弹性限流方案包括内核态和用户态配置两种。二者不可同时使用。

用户态通过Linux内核提供的`CFS bandwidth control`能力实现，在保障整机负载水位安全稳定及不影响其他业务运行的前提下，通过双水位机制允许业务容器自适应调整CPU限制，缓解CPU资源瓶颈，提高业务的运行性能。

内核态通过Linux内核提供的`CPU burst`能力，允许容器短时间内突破其cpu使用限制。内核态配置需要用户手动设置和修改每个pod的burst值的大小，rubik不作自适应调整。

### quotaTurbo 用户态解决方案
用户手动为需要自适应调整CPU限额的业务Pod指定“volcano.sh/quota-turbo="true"”注解，（仅针对限额Pod生效，即yaml中指定CPULimit）。
弹性限流用户态策略根据当前整机CPU负载和容器运行情况定时调整白名单容器的CPU quota，并在启停rubik时自动检验并恢复全部容器的quota值 （本节描述的CPU quota指容器当前的cpu.cfs_quota_us参数）。调整策略包括：
1.  整机CPU负载低于警戒水位时，若白名单容器在当前周期受到CPU压制，则rubik按照压制情况缓慢提升容器CPU quota。单轮容器Quota提升总量最多不超过当前节点总CPU quota的1%。
2.  整机CPU负载高于高水位时，若白名单容器在当前周期未受到CPU压制，则rubik依据水位慢速回调容器quota值。
3.  整机CPU负载高于警戒水位时，若白名单容器当前Quota值超过配置值，则rubik快速回落所有容器CPU quota值，尽力保证负载低于警戒水位。
4.  容器最大可调整CPU quota不超过2倍用户配置值（例如Pod yaml中指定CPUlimit参数），但不应小于用户配置值。
5.  容器在60个同步间隔时间内的整体CPU利用率不得超过用户配置值。
    
### quotaBurst 内核态解决方案
用户手动为需要业务Pod指定“volcano.sh/quota-burst-time”注解，rubik将注解值写入pod的所有容器的burst内核接口，写入后rubik不会根据cpu使用率等自适应修改。示例如下：
```yaml
metadata:    
  annotations:    
    volcano.sh/quota-burst-time : "2000"
```
> 内核态通过内核接口cpu.cfs_burst_us实现。支持内核态配置需要确认cgroup的cpu子系统目录下存在cpu.cfs_burst_us文件，其值约束如下：
> 1. 当cpu.cfs_quota_us的值不为-1时，需满足cfs_burst_us + cfs_quota_us <= 2^44-1 且 cfs_burst_us <= cfs_quota_us。
> 2. 当cpu.cfs_quota_us的值为-1时，CPU burst功能不生效，cfs_burst_us默认为0，不支持配置其他任何值。

### 约束限制
-   用户态通过CFS bandwidth control调整cpu.cfs_period_us和cpu.cfs_quota_us参数实现CPU带宽控制。因此用户态约束如下：
    -  禁止第三方更改CFS bandwidth control相关参数（包括但不限于cpu.cfs_quota_us、cpu.cfs_period_us等文件），以避免未知错误。
    -  禁止与具有限制CPU资源功能的同类产品同时使用，否则导致用户态功能无法正常使用。同类产品包括但不限于：kubernetes的VPA、HPA，腾讯的EVPA，Alibaba的CPU Burst，cgroup提供的cpu-share和绑核功能等。
    -  若用户监控CFS bandwidth control相关指标，使用本特性可能会破坏监测指标的一致性。
-   内核态约束如下：
    -   用户应使用k8s接口设置pod的busrt值，禁止用户手动直接修改容器的cpu cgroup目录下的cpu.cfs_burst_us文件。
- 禁止用户同时使能弹性限流用户态和内核态方案。

## ioCost 支持iocost对IO权重控制
为了有效解决由离线业务IO占用过高，导致在线业务QoS下降的问题，rubik容器提供了基于cgroup v1 iocost的IO权重控制功能。
资料参见：
[iocost内核相关功能介绍]([https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#io:~:text=correct%20memory%20ownership.-,IO,-%C2%B6)

### 约束限制
- kernel需要支持cgroup v1 iocost
- 内核启动参数需要增加cgroup1_writeback，可以通过cat /pro/cmdline确认是否有该启动参数
- rubik容器需要挂载主机的/dev目录，因为rubik会读取硬盘的相关参数。
- iocost只支持配置到物理块设备，例如/dev/sda硬盘等。
- iocost生效基于当时IO资源紧张程度是否达到配置的参数。如果达到配置的参数，则会使用iocost配置的权重进行IO资源分配，因此如果要获取比较好的结果可以适当缩小配置值，但是有可能会影响整体IO的资源使用。