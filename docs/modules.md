# 模块介绍

## Pod CPU优先级

rubik支持业务CPU优先级配置，针对在离线业务混合部署的场景，确保在线业务相对离线业务的CPU资源抢占。

**前置条件**：

- 建议内核版本openEuler-22.03+。内核支持针对cgroup的cpu优先级配置，cpu子系统存在接口cpu.qos_level。

### CPU优先级内核接口

- /sys/fs/cgroup/cpu目录下容器的cgroup中，如`/sys/fs/cgroup/cpu/kubepods/burstable/<PodUID>/<container-longid>`目录
  - cpu.qos_level：开启CPU优先级配置，默认值为0， 有效值为0和-1。
    - 0：标识为在线业务
    - -1：标识为离线业务

### CPU优先级配置详解

rubik会根据pod的yaml文件中的注解`volcano.sh/preemptable`自动配置cpu.qos_level, 默认为false。

```
annotations:
    volcano.sh/preemptable: true
```

- true：代表业务为离线业务，
- false：代表业务为在线业务

---------------------

## pod 内存优先级

rubik支持业务memory优先级配置，针对在离线业务混合部署的场景，确保OOM时优先kill离线业务。

**前置条件**：

- 建议内核版本openEuler-22.03+。内核支持针对cgroup的memory优先级配置，memory子系统存在接口memory.qos_level。
- 开启内存优先级支持: `echo 1 > /proc/sys/vm/memcg_qos_enable`

### 内存优先级内核接口

- /proc/sys/vm/memcg_qos_enable：开启内存优先级特性，默认值为0，有效值为0和1。开启命令为：`echo 1 > /proc/sys/vm/memcg_qos_enable`
  - 0：表示关闭特性
  - 1：表示开启特性。

- /sys/fs/cgroup/memory目录下容器的cgroup中，如`/sys/fs/cgroup/memory/kubepods/burstable/<PodUID>/<container-longid>`目录
  - memory.qos_level：开启内存优先级配置，默认值为0，有效值为0和-1。
    - 0：标识为在线业务
    - -1：标识为离线业务

### 内存优先级配置详解

rubik会根据pod的yaml文件中的注解`volcano.sh/preemptable`自动配置memory.qos_level，参考[CPU优先级配置详解](#cpu优先级配置详解)

---------------------

## dynCache 访存带宽和LLC限制

rubik支持业务的Pod访存带宽(memory bandwidth)和LLC(Last Level Cache)限制，通过限制离线业务的访存带宽/LLC使用，减少其对在线业务的干扰。

**前置条件**：

- cache/访存限制功能仅支持物理机，不支持虚拟机。
  - X86物理机，需要OS支持且开启intel RDT的CAT和MBA功能，内核启动项cmdline需要添加`rdt=l3cat,mba`
  - ARM物理机，需要OS支持且开启mpam功能，内核启动项需要添加`mpam=acpi`。

- 由于内核限制，RDT mode当前不支持pseudo-locksetup模式。

**rubik新增权限和目录**：

- 挂载目录: `/sys/fs/resctrl`。 rubik需要读取和设置/sys/fs/resctrl目录下的文件，该目录需在rubik启动前挂载，且需保障在rubik运行过程中不被卸载。
- 权限: SYS_ADMIN. 设置主机/sys/fs/resctrl目录下的文件需要rubik容器被赋有SYS_ADMIN权限。
- namepsace: pid namespace. rubik需要获取业务容器进程在主机上的pid，所以rubik容器需与主机共享pid namespace。

**rubik rdt 控制组**：

rubik在RDT resctrl目录(默认为 /sys/fs/resctrl)下创建5个控制组，分别为rubik_max、rubik_high、rubik_middle、rubik_low、rubik_dynamic。rubik启动后，将水位线写入对应控制组的schemata。其中，low、middle、high的水位线可在cacheConfig中配置；max控制组为默认最大值, dynamic控制组初始水位线和low控制组一致。

离线业务pod启动时通过注解`volcano.sh/cache-limit`设置其cache level，并被加入到指定的控制组中, 如下列配置的pod将被加入rubik_low控制组:

```
annotations:
    volcano.sh/cache-limit: "low"
```

**rubik dynamic控制组**：

当存在level为dynamic的离线pod时，rubik通过采集当前节点在线业务pod的cache miss 和 llc miss 指标，调整rubik_dynamic控制组的水位线，实现对dynamic控制组内离线应用pod的动态控制。

### dynCache内核接口

- /sys/fs/resctrl: 在该目录下创建5个控制组目录，并修改其schemata和tasks文件。

### dynCache配置详解

dynCache功能相关的配置在`cacheConfig`中:

```
"cacheConfig": {
        "enable": false,
        "defaultLimitMode": "static",
        "adjustInterval": 1000,
        "perfDuration": 1000,
        "l3Percent": {
            "low": 20,
            "mid": 30,
            "high": 50
        },
        "memBandPercent": {
            "low": 10,
            "mid": 30,
            "high": 50
        }
    },
```

- l3Percent 和 memBandPercent:
    通过 l3Percent 和 memBandPercent 配置low, mid, high控制组的水位线。

    比如当环境的`rdt bitmask=fffff, numa=2`时, rubik_low的控制组将根据 l3Percent low=20 和 memBandPercent low=10 两个参数, 将为/sys/fs/resctrl/rubik_low控制组配置:

    ```
    L3:0=f;1=f
    MB:0=10;1=10
    ```

- defaultLimitMode: 如果离线pod未指定`volcano.sh/cache-limit`注解，将根据cacheConfig的defaultLimitMode来决定pod将被加入哪个控制组:
  - defaultLimitMode为static时，pod将被加入到rubik_max控制组
  - defaultLimitMode为dynamic时，pod将被加入到rubik_dynamic控制组
- adjustInterval: dynCache动态调整rubik_dynamic控制组的间隔时间，单位ms，默认1000ms
- perfDuration: dynCache性能perf执行时长，单位ms，默认1000ms

### dynCache注意事项

- dynCache仅针对离线pod，对在线业务不生效。
- 若业务容器运行过程中被手动重启（容器ID不变但容器进程PID变化），针对该容器的dynCache无法生效。
- 业务容器启动并已设置dynCache级别后，不支持对其限制级别进行修改。
- 动态限制组的调控灵敏度受到rubik配置文件内adjustInterval、perfDuration值以及节点在线业务pod数量的影响，每次调整（若干扰检测结果为需要调整）间隔在区间[adjustInterval+perfDuration, adjustInterval+perfDuration*pod数量]内波动，用户可根据灵敏度需求调整配置项。

---------------------

## blkio

Pod的blkio的配置以`volcano.sh/blkio-limit`注解的形式，在pod创建的时候配置，或者在pod运行期间通过kubectl annotate进行动态的修改，支持离线和在线pod。

配置内容为4个列表:
| 项                | 说明                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| device_read_bps   | 用于设定设备执行“读”操作字节的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备，value限定上限值，单位为byte   |
| device_read_iops  | 用于设定设备执行“读”操作次数的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备                                |
| device_write_bps  | 用于设定设备执行 “写” 操作次数的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备，value限定上限值，单位为byte |
| device_write_iops | 用于设定设备执行“写”操作字节的上限。该配置为list，可以对多个device进行配置，device指定需要限制的块设备                                |

### blkio内核接口

- /sys/fs/cgroup/blkio目录下容器的cgroup中，如`/sys/fs/cgroup/blkio/kubepods/burstable/<PodUID>/<container-longid>`目录:
  - blkio.throttle.read_bps_device
  - blkio.throttle.read_iops_device
  - blkio.throttle.write_bps_device
  - blkio.throttle.write_iops_device

配置的key:value和cgroup的key:value的配置规则一致:

- 写入时会转换成环境page size的倍数
- 只有minor为0的device配置才会生效
- 如果取消限速，可将值设为0

### blkio配置详解

**rubik开启关闭blkio功能**:

rubik提供blkio配置功能的开关，在`blkConfig`中

```
"blkConfig": {
        "limit": true
}
```

- limit: IO控制模块使能开关， 默认为false

**pod配置样例**:

通过pod的注解配置时可提供四个列表，分别是write_bps, write_iops, read_bps, read_iops, read_byte.

- 创建时: 在yaml文件中

  ```
  volcano.sh/blkio-limit: '{"device_read_bps":[{"device":"/dev/sda1","value":"10485760"}, {"device":"/dev/sda","value":"20971520"}],
                  "device_write_bps":[{"device":"/dev/sda1","value":"20971520"}],
                  "device_read_iops":[{"device":"/dev/sda1","value":"200"}],
                  "device_write_iops":[{"device":"/dev/sda1","value":"300"}]}'
  ```

- 修改annotation: 可通过 kubectl annotate动态修改，如:
  ```kubectl annotate --overwrite pods <podname> volcano.sh/blkio-limit='{"device_read_bps":[{"device":"/dev/vda", "value":"211715200"}]}'```

---------------------

## memory

rubik中支持多种内存策略。针对不同场景使用不同的内存分配方案,以解决多场景内存分配。

dynlevel策略：基于内核cgroup的多级别控制。通过监测节点内存压力，多级别动态调整离线业务的memory cgroup，尽可能地保障在线业务服务质量。

### memory dynlevel策略内核接口

- /sys/fs/cgroup/memory目录下容器的cgroup中，如`/sys/fs/cgroup/memory/kubepods/burstable/<PodUID>/<container-longid>`目录。dynlevel策略会依据当前节点的内存压力大小，依次调整节点离线应用容器的下列值:

  - memory.soft_limit_in_bytes
  - memory.force_empty
  - memory.limit_in_bytes
  - /proc/sys/vm/drop_caches

### memory dynlevel策略配置详解

rubik提供memory的指定策略和控制间隔，在`memConfig`中

```
"memConfig": {
        "strategy": "none",
        "checkInterval": 5
   }
```

- strategy为memory的策略名称，现支持 dynlevel 和 none 两个选项，默认为none。
  - none: 即不设置任何策略，不会对内存进行调整。
  - dynlevel: 动态分级调整策略。

- checkInterval为策略的周期性检查的时间，单位为秒, 默认为5。

---------------------

## quota burst

Pod的quota burst的配置以`volcano.sh/quota-burst-time`注解的形式，在pod创建的时候配置，或者在pod运行期间通过kubectl annotate进行动态的修改，支持离线和在线pod。

Pod的quota burst默认单位是microseconds, 其允许容器的cpu使用率低于quota时累积cpu资源，并在cpu利用率超过quota时，使用容器累积的cpu资源。

### quota burst内核接口

- /sys/fs/cgroup/cpu目录下容器的cgroup中，如`/sys/fs/cgroup/cpu/kubepods/burstable/<PodUID>/<container-longid>`目录，注解的值将被写入下列文件中：
  - cpu.cfs_burst_us

- 注解`volcano.sh/quota-burst-time`的值和cpu.cfs_burst_us的约束一致：
  - 当cpu.cfs_quota_us不为-1，需满足cpu.cfs_burst_us + cpu.cfs_quota_us <= 2^44-1 且 cpu.cfs_burst_us <= cpu.cfs_quota_us
  - 当cpu.cfs_quota_us为-1，cpu.cfs_burst_us最大没有限制，取决于系统最大可设置的值


**pod配置样例**

- 创建时: 在yaml文件中

  ```
  metadata:
    annotations:
      volcano.sh/quota-burst-time : "2000"
  ```

- 修改annotation: 可通过 kubectl annotate动态修改，如:

  ```kubectl annotate --overwrite pods <podname> volcano.sh/quota-burst-time='3000'```
