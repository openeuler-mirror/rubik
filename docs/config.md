# 配置

配置文件内容：

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

配置说明：

- autoConfig: bool类型。

  合法值：false、true。

  false表示关闭Pod自动感知配置功能。

  true表示开启Pod自动感知配置功能。

- autoCheck：bool类型。

  合法值：false、true。

  false表示关闭Pod优先级校验功能。

  true表示开启Pod优先级校验功能。

- logDriver：string类型。

  合法值：stdio、file。

  stdio表示直接向标准输出打印日志，日志收集和转储由调度平台完成。

  file表示将文件打印到日志目录。

- logDir：string类型。

  日志目录。

- logSize：int类型。

  指定日志大小，单位MB；范围(1MB, 1TB)。

- logLevel：string类型。

  日志级别，合法值：debug, info, error。

- cgroupRoot：string类型。

  指定cgroup挂载点。