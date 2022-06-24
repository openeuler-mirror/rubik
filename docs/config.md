# Rubik配置说明

## 基本配置说明

Rubik执行程序由Go语言实现，并编译为静态可执行文件，以便尽可能与系统依赖解耦。

Rubik除支持 `-v` 参数查询版本信息之外，不支持其他参数，版本信息输出示例如下所示，该信息中的内容和格式可能随着版本发生变化。

```
rubik -v
Version:       0.1.0
Release:
Go Version:    go1.15.15
Git Commit:    9bafc90
Built:         2022-06-24
OS/Arch:       linux/amd64
```

Rubik启动时会解析配置文件，配置文件的路径固定为 `/var/lib/rubik/config.json` ，为避免配置混乱，暂不支持指定其他路径。

配置文件采用json格式，字段键采用驼峰命名规则，且首字母小写。

配置文件示例内容如下：

```json
{
    "autoConfig": true,
    "autoCheck": false,
    "logDriver": "stdio",
    "logDir": "/var/log/rubik",
    "logSize": 1024,
    "logLevel": "info",
    "cgroupRoot": "/sys/fs/cgroup"
}
```

常用配置项说明：

| 配置键[=默认值]           | 类型   | 描述                                                | 示例值               |
| ------------------------- | ------ | --------------------------------------------------- | -------------------- |
| autoConfig=false          | bool   | 自动配置开关，自动配置即自行拉取Pod信息并配置给系统 | false, true          |
| autoCheck=false           | bool   | 自动检查开关，自动纠正因故障等原因导致的错误配置    | false, true          |
| logDriver=stdio           | string | 日志驱动，支持标准输出和文件                        | stdio, file          |
| logDir=/var/log/rubik     | string | 日志保存目录                                        | /var/log/rubik       |
| logSize=1024              | int    | 总日志大小，单位MB，适用于logDriver=file            | [10, 2**20]          |
| logLevel=info             | string | 日志级别                                            | debug, info, error   |
| cgroupRoot=/sys/fs/cgroup | string | 系统cgroup挂载点路径                                | /sys/fs/cgroup       |
| cacheConfig               | map    | 动态控制CPU高速缓存模块（dynCache）的相关配置       |                      |
| enable=false              | bool   | dynCache功能启用开关                                | false, true          |
| defaultLimitMode=static   | string | dynCache控制模式                                    | static, dynamic      |
| adjustInterval=1000       | int    | dynCache动态控制间隔时间，单位ms                    | [10, 10000]          |
| perfDuration=1000         | int    | dynCache性能perf执行时长，单位ms                    | [10, 10000]          |
| l3Percent                 | map    | dynCache控制中L3各级别对应水位（%）                 |                      |
| low=20                    | int    | L3低水位组控制线                                    | [1, 100]             |
| mid=30                    | int    | L3中水位组控制线                                    | [1, 100]             |
| high=50                   | int    | L3高水位组控制线                                    | [1, 100]             |
| memBandPercent            | map    | dynCache控制中MB各级别对应水位（%）                 |                      |
| low=10                    | int    | MB低水位组控制线                                    | [1, 100]             |
| mid=30                    | int    | MB中水位组控制线                                    | [1, 100]             |
| high=50                   | int    | MB高水位组控制线                                    | [1, 100]             |
| blkioConfig               | map    | IO控制模块相关配置                                  |                      |
| limit=false               | bool   | IO控制模块使能开关                                  |                      |
| memConfig                 | map    | 内存控制模块相关配置                                |                      |
| strategy=none             | string | 内存动态分级回收控制策略                            | none, dynlevel, fssr |



