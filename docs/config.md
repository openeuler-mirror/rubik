# Rubik配置说明

Rubik执行程序由Go语言实现，并编译为静态可执行文件，以便尽可能与系统依赖解耦。

## 命令

Rubik仅支持 使用`-v` 参数查询版本信息，不支持其他参数。
版本信息输出示例如下所示，该信息中的内容和格式可能随着版本发生变化。

```bash
$ ./rubik -v
Version:       2.0.0
Release:       NA
Go Version:    go1.18.8
Git Commit:    bcaace8
Built:         2023-03-30
OS/Arch:       linux/amd64
```

## 配置

执行rubik二进制时，rubik首先会解析配置文件，配置文件的路径固定为 `/var/lib/rubik/config.json`。

> 1. 为避免配置混乱，暂不支持指定其他路径。
> 2. rubik支持以daemonset形式运行在kubernetes集群中。我们提供了yaml脚本（`hack/rubik-daemonset.yaml`），并定义了`ConfigMap`作为配置。

因此，以daemonset形式运行rubik时，应修改`hack/rubik-daemonset.yaml`中的相应配置。

配置文件采用json格式，字段键采用驼峰命名规则，且首字母小写。
配置文件示例内容如下：

```json
{
  "agent": {
    "logDriver": "stdio",
    "logDir": "/var/log/rubik",
    "logSize": 2048,
    "logLevel": "info",
    "cgroupRoot": "/sys/fs/cgroup",
    "informerType": "apiserver",
    "cgroupDriver": "cgroupfs",
    "enabledFeatures": [
      "preemption",
      "dynCache",
      "ioLimit",
      "ioCost",
      "quotaBurst",
      "quotaTurbo"
    ]
  },
  "preemption": {
    "resource": [
      "cpu",
      "memory"
    ]
  },
  "quotaTurbo": {
    "highWaterMark": 50,
    "syncInterval": 100
  },
  "dynCache": {
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
  "ioCost": [
    {
      "nodeName": "k8s-single",
      "config": [
        {
          "dev": "sdb",
          "enable": true,
          "model": "linear",
          "param": {
            "rbps": 10000000,
            "rseqiops": 10000000,
            "rrandiops": 10000000,
            "wbps": 10000000,
            "wseqiops": 10000000,
            "wrandiops": 10000000
          }
        }
      ]
    }
  ]
}
```

Rubik配置分为两类：通用配置和特性配置。

- 通用配置由agent关键字标识，用于保存全局的配置。
- 特性配置按服务类型区分，应用于各个子特性。特性配置必须在通用配置的`enabledFeatures`字段中声明方可使能。

### agent

`agent`配置用于记录保存rubik运行的通用配置，例如日志、cgroup挂载点、cgroup驱动等信息。
| 配置键[=默认值]           | 类型       | 描述                                   | 可选值                      |
| ------------------------- | ---------- | ------------------------------------ | ---------------------------|
| logDriver=stdio           | string     | 日志驱动，支持标准输出和文件              | stdio, file                 |
| logDir=/var/log/rubik     | string     | 日志保存目录                           | 可读可写的目录                |
| logSize=1024              | int        | 日志限额，单位MB，仅logDriver=file生效   | [10, $2^{20}$]              |
| logLevel=info             | string     | 输出日志级别                           | debug,info,warn,error       |
| cgroupRoot=/sys/fs/cgroup | string     | 系统cgroup挂载点路径                    | 系统cgroup挂载点路径          |
| cgroupDriver=cgroupfs     | string     | cgroup驱动类型                         | cgroupfs、systemd           |
| enabledFeatures=[]        | string数组 | 需要使能的rubik特性列表                 | rubik支持特性，参见特性介绍     |
| informerType=apiserver    | string     | informer类型                          | apiserver、nri              |

#### informerType

- apiserver（默认方式）。rubik通过list-watch机制从kubernetes apiserver中获取pod和容器数据。
- nri。rubik通过nri套接字从容器引擎中获取数据，nri套接字路径固定为`/var/run/nri/nri.sock`。若rubik运行在容器中，需将nri套接字挂载到容器中。

挂载套接字的配置文件如下：

```yaml
spec:
  # ...
  containers:
    # ...
    volumeMounts:
    - name: nrisock
      mountPath: /var/run/nri/nri.sock
      readOnly: true
  volumes:
    # ...
  - name: nrisock
    hostPath:
      path: /var/run/nri/nri.sock
```

### preemption

`preemption`字段用于标识绝对抢占特性配置。目前，Preemption特性支持CPU和内存的绝对抢占，用户可以按需配置该字段，单独或组合使用资源的绝对抢占。
| 配置键[=默认值] | 类型       | 描述                             | 可选值      |
| --------------- | ---------- | -------------------------------- | ----------- |
| resource=[]     | string数组 | 资源类型，声明何种资源需要被访问 | cpu, memory |

### dynCache

`dynCache`字段用于标识支持Pod内存带宽和LLC限制特性配置。`l3Percent`字段用于标识最后一级缓存（LLC）水位控制线，`memBandPercent`字段用于标识内存带宽（MB）水位控制线。

| 配置键[=默认值]         | 类型   | 描述               | 可选值          |
| ----------------------- | ------ | ------------------ | --------------- |
| defaultLimitMode=static | string | dynCache的控制模式 | static, dynamic |
| adjustInterval=1000      | int    | dynCache动态控制间隔时间，单位ms| [10, 10000] |
| perfDuration=1000        | int    | dynCache性能perf执行时长，单位ms | [10, 10000] |
| l3Percent                | map    | dynCache控制中L3各级别对应水位（%）|      |
| .low=20                  | int    | L3 Cache低水位组控制线 | [10, 100]     |
| .mid=30                  | int    | L3 Cache中水位组控制线  | [low, 100]   |
| .high=50                 | int    | L3 Cache高水位组控制线  | [mid, 100]   |
| memBandPercent           | map    | dynCache控制中MB各级别对应水位（%）|   |
| .low=10                  | int    | MB（内存带宽）低水位组控制线 | [10, 100]  |
| .mid=30                  | int    | MB中水位组控制线  | [low, 100]   |
| .high=50                 | int    | MB高水位组控制线 | [mid, 100]   |

### quotaTurbo

`quotaTurbo`字段用于标识支持弹性限流技术（用户态）配置。
| 配置键[=默认值]   | 类型   | 描述                             | 可选值               |
| ----------------- | ------ | -------------------------------- | -------------------- |
| highWaterMark=60  | int | CPU负载的高水位值         |\[0,警戒水位) |
| alarmWaterMark=80 | int | CPU负载的警戒水位 | (高水位,100\]            |
| syncInterval=100  | int | 触发容器quota值更新的间隔（单位：毫秒） | [100,10000] |

### ioCost

`ioCost`字段用于标识支持iocost对IO权重控制特性配置。其类型为数组，数组中的每一个元素由节点名称`nodeName`和设备参数数组`config`组成。
| 配置键  | 类型   | 描述                             | 可选值               |
| ----------------- | ------ | -------------------------------- | -------------------- |
| nodeName  | string | 节点名称         | kubernetes中节点名称 |
| config | 数组 | 单个设备的配置信息 |   /       |

单个块设备配置`config`参数：
| 配置键[=默认值] | 类型   | 描述                                          | 可选值         |
| --------------- | ------ | --------------------------------------------- | -------------- |
| dev             | string | 块设备名称，仅支持物理设备                    | /              |
| model           | string | iocost模型名                                | linear         |
| param           | /      | 设备参数，根据不同模型有不同参数               | /              |

模型为linear时，`param`字段支持如下参数：
| 配置键[=默认值] | 类型 | 描述 | 可选值 |
| --------------- | ---- | ---- | ------ |
|rbps            | int64  | 块设备最大读带宽     | （0, $2^{63}$) |
| rseqiops        | int64  | 块设备最大顺序读iop  | （0, $2^{63}$) |
| rrandiops       | int64  | 块设备最大随机读iops | （0, $2^{63}$) |
| wbps            | int64  | 块设备最大写带宽     | （0, $2^{63}$) |
| wseqiops        | int64  | 块设备最大顺序写iops | （0, $2^{63}$) |
| wrandiops       | int64  | 块设备最大随机写iops | （0, $2^{63}$) |
